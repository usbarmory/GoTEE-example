// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package util

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// Console represents an SSH console instance.
type Console struct {
	// Banner is the login welcome banner
	Banner string
	// Help is the `help` command output
	Help string
	// Handler is the terminal command handler
	Handler func(*terminal.Terminal, string) error
	// Term is the terminal instance
	Term *terminal.Terminal
}

func (c *Console) handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		_ = newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	conn, requests, err := newChannel.Accept()

	if err != nil {
		log.Printf("error accepting channel, %v", err)
		return
	}

	c.Term = terminal.NewTerminal(conn, "")
	c.Term.SetPrompt(string(c.Term.Escape.Red) + "> " + string(c.Term.Escape.Reset))

	go func() {
		defer conn.Close()

		log.SetOutput(io.MultiWriter(log.Writer(), c.Term))
		defer log.SetOutput(log.Writer())

		fmt.Fprintf(c.Term, "%s\n", c.Banner)
		fmt.Fprintf(c.Term, "%s\n", string(c.Term.Escape.Cyan)+c.Help+string(c.Term.Escape.Reset))

		for {
			cmd, err := c.Term.ReadLine()

			if err == io.EOF {
				break
			}

			if err != nil {
				log.Printf("readline error: %v", err)
				continue
			}

			err = c.Handler(c.Term, cmd)

			if err == io.EOF {
				break
			}

			if err != nil {
				log.Printf("error: %v", err)
			}
		}

		log.Printf("closing ssh connection")
	}()

	go func() {
		for req := range requests {
			reqSize := len(req.Payload)

			switch req.Type {
			case "shell":
				// do not accept payload commands
				if len(req.Payload) == 0 {
					_ = req.Reply(true, nil)
				}
			case "pty-req":
				// p10, 6.2.  Requesting a Pseudo-Terminal, RFC4254
				if reqSize < 4 {
					log.Printf("malformed pty-req request")
					continue
				}

				termVariableSize := int(req.Payload[3])

				if reqSize < 4+termVariableSize+8 {
					log.Printf("malformed pty-req request")
					continue
				}

				w := binary.BigEndian.Uint32(req.Payload[4+termVariableSize:])
				h := binary.BigEndian.Uint32(req.Payload[4+termVariableSize+4:])

				_ = c.Term.SetSize(int(w), int(h))
				_ = req.Reply(true, nil)
			case "window-change":
				// p10, 6.7.  Window Dimension Change Message, RFC4254
				if reqSize < 8 {
					log.Printf("malformed window-change request")
					continue
				}

				w := binary.BigEndian.Uint32(req.Payload)
				h := binary.BigEndian.Uint32(req.Payload[4:])

				_ = c.Term.SetSize(int(w), int(h))
			}
		}
	}()
}

func (c *Console) handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go c.handleChannel(newChannel)
	}
}

func (c *Console) listen(listener net.Listener, srv *ssh.ServerConfig) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Printf("error accepting connection, %v", err)
			continue
		}

		sshConn, chans, reqs, err := ssh.NewServerConn(conn, srv)

		if err != nil {
			log.Printf("error accepting handshake, %v", err)
			continue
		}

		log.Printf("new ssh connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())

		go ssh.DiscardRequests(reqs)
		go c.handleChannels(chans)
	}
}

// Start instantiates an SSH console on the given listener.
func (c *Console) Start(listener net.Listener) (err error) {
	srv := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		return fmt.Errorf("private key generation error: ", err)
	}

	signer, err := ssh.NewSignerFromKey(key)

	if err != nil {
		return fmt.Errorf("key conversion error: ", err)
	}

	log.Printf("starting ssh server (%s)", ssh.FingerprintSHA256(signer.PublicKey()))

	srv.AddHostKey(signer)

	go c.listen(listener, srv)

	return
}

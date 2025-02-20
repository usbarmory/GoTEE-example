// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"io"
	"regexp"
	"runtime/debug"
	"runtime/pprof"

	"golang.org/x/term"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
)

func init() {
	Add(Cmd{
		Name: "help",
		Help: "this help",
		Fn:   helpCmd,
	})

	Add(Cmd{
		Name:    "exit, quit",
		Args:    1,
		Pattern: regexp.MustCompile(`^(exit|quit)$`),
		Help:    "close session",
		Fn:      exitCmd,
	})

	Add(Cmd{
		Name: "stack",
		Help: "stack trace of current goroutine",
		Fn:   stackCmd,
	})

	Add(Cmd{
		Name: "stackall",
		Help: "stack trace of all goroutines",
		Fn:   stackallCmd,
	})

	Add(Cmd{
		Name: "reboot",
		Help: "reset device",
		Fn:   rebootCmd,
	})
}

func helpCmd(term *term.Terminal, _ []string) (string, error) {
	return Help(term), nil
}

func exitCmd(_ *term.Terminal, _ []string) (string, error) {
	return "logout", io.EOF
}

func stackCmd(_ *term.Terminal, _ []string) (string, error) {
	return string(debug.Stack()), nil
}

func stackallCmd(_ *term.Terminal, _ []string) (string, error) {
	buf := new(bytes.Buffer)
	pprof.Lookup("goroutine").WriteTo(buf, 1)

	return buf.String(), nil
}

func rebootCmd(_ *term.Terminal, _ []string) (_ string, _ error) {
	usbarmory.Reset()
	return
}

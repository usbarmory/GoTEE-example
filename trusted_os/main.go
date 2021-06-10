// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
	_ "unsafe"

	"github.com/f-secure-foundry/tamago/board/f-secure/usbarmory/mark-two"
	"github.com/f-secure-foundry/tamago/dma"
	"github.com/f-secure-foundry/tamago/soc/imx6"
	"github.com/f-secure-foundry/tamago/soc/imx6/dcp"
	"github.com/f-secure-foundry/tamago/soc/imx6/usb"

	"github.com/f-secure-foundry/imx-usbnet"

	"github.com/f-secure-foundry/GoTEE-example/mem"
	"github.com/f-secure-foundry/GoTEE-example/util"
)

const (
	sshPort   = 22
	deviceIP  = "10.0.0.1"
	deviceMAC = "1a:55:89:a2:69:41"
	hostMAC   = "1a:55:89:a2:69:42"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = mem.SecureStart

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = mem.SecureSize

var ssh *util.Console

func init() {
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout)

	if imx6.Native {
		if err := imx6.SetARMFreq(900); err != nil {
			panic(fmt.Sprintf("WARNING: error setting ARM frequency: %v", err))
		}

		debugConsole, _ := usbarmory.DetectDebugAccessory(250 * time.Millisecond)
		<-debugConsole
	}

	// Move DMA region prevent free NonSecure access, alternatively
	// iRAM/OCRAM (default DMA region) can be locked down.
	//
	// This is necessary as iRAM/OCRAM is outside TZASC control.
	dma.Init(mem.SecureDMAStart, mem.SecureDMASize)

	log.Printf("PL1 %s/%s (%s) • TEE system/monitor (Secure World)", runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func gotee() {
	var wg sync.WaitGroup

	ta := loadApplet()
	os := loadNormalWorld(false)

	// test concurrent execution of:
	//   Secure    World PL1 (system/monitor mode) - secure OS (this program)
	//   Secure    World PL0 (user mode)           - trusted applet
	//   NonSecure World PL1                       - main OS
	wg.Add(2)
	go run(ta, &wg)
	go run(os, &wg)

	if !imx6.Native {
		go func() {
			for i := 0; i < 60; i++ {
				time.Sleep(1 * time.Second)
				log.Printf("PL1 says %d missisipi", i+1)
			}
		}()
	}

	log.Printf("PL1 waiting for applet and kernel")
	wg.Wait()

	usbarmory.LED("blue", false)

	if imx6.Native {
		// re-launch NonSecure World with peripheral restrictions
		os = loadNormalWorld(true)

		log.Printf("PL1 re-launching kernel with TrustZone restrictions")
		run(os, nil)

		// test restricted peripheral in Secure World
		log.Printf("PL1 in Secure World is about to perform DCP key derivation")

		k, err := dcp.DeriveKey(make([]byte, 8), make([]byte, 16), -1)

		if err != nil {
			log.Printf("PL1 in Secure World World failed to use DCP (%v)", err)
		} else {
			log.Printf("PL1 in Secure World World successfully used DCP (%x)", k)
		}
	}
}

func main() {
	defer log.Printf("PL1 says goodbye")

	if !imx6.Native {
		gotee()
		return
	}

	gonet, err := usbnet.Init(deviceIP, deviceMAC, hostMAC, 1)

	if err != nil {
		log.Fatalf("PL1 could not initialize USB networking, %v", err)
	}

	gonet.EnableICMP()

	listener, err := gonet.ListenerTCP4(sshPort)

	if err != nil {
		log.Fatalf("PL1 could not initialize SSH listener, %v", err)
	}

	ssh = &util.Console{
		Banner:  fmt.Sprintf("PL1 %s/%s (%s) • TEE system/monitor (Secure World)", runtime.GOOS, runtime.GOARCH, runtime.Version()),
		Help:    help,
		Handler: cmd,
		Listener: listener,
	}

	err = ssh.Start()

	if err != nil {
		log.Fatalf("PL1 could not initialize SSH server, %v", err)
	}

	usb.USB1.Init()
	usb.USB1.DeviceMode()
	usb.USB1.Reset()

	// never returns
	usb.USB1.Start(gonet.Device())
}

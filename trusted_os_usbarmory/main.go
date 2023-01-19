// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
	_ "unsafe"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"

	"github.com/usbarmory/imx-usbnet"

	"github.com/usbarmory/GoTEE/monitor"

	"github.com/usbarmory/GoTEE-example/mem"
	"github.com/usbarmory/GoTEE-example/util"
)

const (
	sshPort = 22
	IP      = "10.0.0.1"
	MAC     = "1a:55:89:a2:69:41"
	hostMAC = "1a:55:89:a2:69:42"
)

// TrustZone Watchdog interval (in ms) to force Non-Secure to Secure World
// switching.
const watchdogTimeout = 10000

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = mem.SecureStart

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = mem.SecureSize

var ssh *util.Console

func init() {
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout)

	// Move DMA region to prevent NonSecure access, alternatively
	// iRAM/OCRAM (default DMA region) can be locked down on its own (as it
	// is outside TZASC control).
	dma.Init(mem.SecureDMAStart, mem.SecureDMASize)

	if imx6ul.Native {
		imx6ul.SetARMFreq(900)

		imx6ul.DCP.Init()
		imx6ul.DCP.DeriveKeyMemory = dma.Default()

		debugConsole, _ := usbarmory.DetectDebugAccessory(250 * time.Millisecond)
		<-debugConsole
	}

	log.Printf("%s/%s (%s) • TEE security monitor (Secure World system/monitor)", runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func gotee() (err error) {
	var ta *monitor.ExecCtx
	var os *monitor.ExecCtx
	var wg sync.WaitGroup

	if ta, err = loadApplet(); err != nil {
		return
	}

	if os, err = loadNormalWorld(false); err != nil {
		return
	}

	// test concurrent execution of:
	//   Secure    World PL1 (system/monitor mode) - secure OS (this program)
	//   Secure    World PL0 (user mode)           - trusted applet
	//   NonSecure World PL1                       - main OS
	wg.Add(2)
	go run(ta, &wg)
	go run(os, &wg)

	if !imx6ul.Native {
		go func() {
			for i := 0; i < 60; i++ {
				time.Sleep(1 * time.Second)
				log.Printf("SM says %d missisipi", i+1)
			}
		}()
	}

	log.Printf("SM waiting for applet and kernel")
	wg.Wait()

	usbarmory.LED("blue", false)

	if !imx6ul.Native {
		return
	}

	// re-launch Normal World with peripheral restrictions
	if os, err = loadNormalWorld(true); err != nil {
		return
	}

	log.Printf("SM re-launching kernel with TrustZone restrictions")
	run(os, nil)

	// test restricted peripheral in Secure World
	log.Printf("SM in Secure World is about to perform DCP key derivation")

	k, err := imx6ul.DCP.DeriveKey(make([]byte, 8), make([]byte, 16), -1)

	if err != nil {
		log.Printf("SM in Secure World failed to use DCP (%v)", err)
	} else {
		log.Printf("SM in Secure World successfully used DCP (%x)", k)
	}

	return
}

func linux(device string) (err error) {
	var os *monitor.ExecCtx

	if os, err = loadLinux(device); err != nil {
		return
	}

	log.Printf("SM enabling TrustZone Watchdog")
	enableTrustZoneWatchdog()

	log.Printf("SM launching Linux")
	run(os, nil)

	return
}

func main() {
	defer log.Printf("SM says goodbye")

	if !imx6ul.Native {
		if err := gotee(); err != nil {
			log.Fatal(err)
		}

		return
	}

	iface, err := usbnet.Init(IP, MAC, hostMAC, 1)

	if err != nil {
		log.Fatalf("SM could not initialize USB networking, %v", err)
	}

	iface.EnableICMP()

	listener, err := iface.ListenerTCP4(sshPort)

	if err != nil {
		log.Fatalf("SM could not initialize SSH listener, %v", err)
	}

	ssh = &util.Console{
		Banner:   fmt.Sprintf("%s/%s (%s) • TEE security monitor (Secure World system/monitor)", runtime.GOOS, runtime.GOARCH, runtime.Version()),
		Help:     help,
		Handler:  cmd,
		Listener: listener,
	}

	err = ssh.Start()

	if err != nil {
		log.Fatalf("SM could not initialize SSH server, %v", err)
	}

	usbarmory.USB1.Init()
	usbarmory.USB1.DeviceMode()
	usbarmory.USB1.Reset()

	// never returns
	usbarmory.USB1.Start(iface.NIC.Device)
}

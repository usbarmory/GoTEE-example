// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"log"
	"os"
	"runtime"
	"sync"
	"time"
	_ "unsafe"

	_ "github.com/usbarmory/tamago/board/qemu/sifive_u"
	"github.com/usbarmory/tamago/dma"

	"github.com/usbarmory/GoTEE/monitor"

	"github.com/usbarmory/GoTEE-example/mem"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint64 = mem.SecureStart

//go:linkname ramSize runtime.ramSize
var ramSize uint64 = mem.SecureSize

func init() {
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout)

	dma.Init(mem.SecureDMAStart, mem.SecureDMASize)

	log.Printf("%s/%s (%s) • TEE Security Monitor (M-mode)", runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func gotee() (err error) {
	var ta *monitor.ExecCtx
	var os *monitor.ExecCtx
	var wg sync.WaitGroup

	if ta, err = loadApplet(); err != nil {
		return
	}

	if os, err = loadSupervisor(); err != nil {
		return
	}

	// test concurrent execution of:
	//   Security Monitor (machine mode)     - secure OS (this program)
	//   Applet (supervisor/user mode)       - trusted applet
	//   Untrusted OS (supervisor/user mode) - main OS
	wg.Add(2)
	go run(ta, &wg)
	go run(os, &wg)

	go func() {
		for i := 0; i < 60; i++ {
			time.Sleep(1 * time.Second)
			log.Printf("SM says %d missisipi", i+1)
		}
	}()

	log.Printf("SM waiting for applet and kernel")
	wg.Wait()

	return
}

func main() {
	defer log.Printf("SM says goodbye")

	if err := gotee(); err != nil {
		log.Fatal(err)
	}
}

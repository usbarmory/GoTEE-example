// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gotee

import (
	"log"
	"sync"
	"time"

	_ "github.com/usbarmory/tamago/board/qemu/sifive_u"

	"github.com/usbarmory/GoTEE/monitor"
)

func GoTEE() (err error) {
	var wg sync.WaitGroup
	var ta *monitor.ExecCtx
	var os *monitor.ExecCtx

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

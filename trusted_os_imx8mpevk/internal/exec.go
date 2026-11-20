// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gotee

import (
	"log"
	"sync"

	_ "github.com/usbarmory/tamago/board/nxp/imx8mpevk"

	"github.com/usbarmory/GoTEE/monitor"
)

func GoTEE() (err error) {
	var wg sync.WaitGroup
	var ta *monitor.ExecCtx

	if ta, err = loadApplet(); err != nil {
		return
	}

	// test concurrent execution of EL1 OS (this program) and EL0 applet
	// (user mode).
	wg.Add(1)
	go run(ta, &wg)

	log.Printf("SM waiting for applet")
	wg.Wait()

	return
}

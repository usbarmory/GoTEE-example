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
	//var os *monitor.ExecCtx

	if ta, err = loadApplet(); err != nil {
		return
	}

	//if os, err = loadNormalWorld(false); err != nil {
	//	return
	//}

	// test concurrent execution of:
	//   Secure    World EL1 (system/monitor mode) - secure OS (this program) - FIXME:linkcpuinit
	//   Secure    World EL0 (user mode)           - trusted applet
	//   NonSecure World EL1                       - main OS (TODO)
	wg.Add(1)
	go run(ta, &wg)
	//go run(os, &wg)

	log.Printf("SM waiting for applet")
	wg.Wait()

	return
}

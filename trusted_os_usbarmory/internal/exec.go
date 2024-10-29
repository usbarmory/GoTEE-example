// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gotee

import (
	"crypto/aes"
	"crypto/sha256"
	"errors"
	"log"
	"math/rand"
	"sync"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"

	"github.com/usbarmory/GoTEE/monitor"
)

// TrustZone Watchdog interval (in ms) to force Non-Secure to Secure World
// switching.
const (
	watchdogTimeout         = 10000
	watchdogWarningInterval = 2000
)

var nsBoot bool

func GoTEE() (err error) {
	var wg sync.WaitGroup
	var ta *monitor.ExecCtx
	var os *monitor.ExecCtx

	if ta, err = loadApplet(false); err != nil {
		return
	}

	if os, err = loadNormalWorld(false); err != nil {
		return
	}

	nsBoot = true

	// test concurrent execution of:
	//   Secure    World PL1 (system/monitor mode) - secure OS (this program)
	//   Secure    World PL0 (user mode)           - trusted applet
	//   NonSecure World PL1                       - main OS
	wg.Add(2)
	go run(ta, &wg)
	go run(os, &wg)

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
	log.Printf("SM in Secure World is about to perform key derivation")

	var k []byte

	switch {
	case imx6ul.CAAM != nil:
		// set CAAM as Secure
		imx6ul.CAAM.SetOwner(true)

		k = make([]byte, sha256.Size)
		err = imx6ul.CAAM.DeriveKey(make([]byte, sha256.Size), k)
	case imx6ul.DCP != nil:
		k, err = imx6ul.DCP.DeriveKey(make([]byte, aes.BlockSize), make([]byte, aes.BlockSize), -1)
	}

	if err != nil {
		log.Printf("SM in Secure World failed to derive key (%v)", err)
	} else {
		log.Printf("SM in Secure World successfully derived key (%x)", k)
	}

	return
}

func Linux(device string) (err error) {
	var os *monitor.ExecCtx

	if nsBoot {
		return errors.New("previous Non-secure kernel run detected, reboot first to launch Linux")
	}

	if os, err = loadLinux(device); err != nil {
		return
	}

	log.Printf("SM enabling TrustZone Watchdog")
	enableTrustZoneWatchdog()

	log.Printf("SM launching Linux")
	run(os, nil)

	return
}

func fault(ctx *monitor.ExecCtx, faultPercentage float64) {
	if n := rand.Float64() * 100; n >= faultPercentage {
		return
	}

	log.Printf("!! injecting register fault !!")
	ctx.R0 += 1
}

func Lockstep(faultPercentage float64) (err error) {
	var once sync.Once
	var ta *monitor.ExecCtx

	if ta, err = loadApplet(true); err != nil {
		return
	}

	defer run(ta, nil)

	if faultPercentage <= 0 {
		return
	}

	primaryHandler := ta.Handler

	ta.Handler = func(ctx *monitor.ExecCtx) error {
		once.Do(func() {
			ta.Shadow.Handler = func(ctx *monitor.ExecCtx) error {
				fault(ctx, faultPercentage)
				return nil
			}
		})

		return primaryHandler(ctx)
	}

	return
}

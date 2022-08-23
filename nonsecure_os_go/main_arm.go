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
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/nxp/imx6ul"

	"github.com/usbarmory/GoTEE-example/mem"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = mem.NonSecureStart

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = mem.NonSecureSize

//go:linkname hwinit runtime.hwinit
func hwinit() {
	imx6ul.Init()
	imx6ul.UART2.Init()
}

//go:linkname printk runtime.printk
func printk(c byte) {
	if imx6ul.Native {
		// monitor call to request logs on Secure World SSH console
		printSecure(c)
	} else {
		imx6ul.UART2.Tx(c)
	}
}

func init() {
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout)

	imx6ul.SetARMFreq(900)
}

func main() {
	log.Printf("%s/%s (%s) • system/supervisor (Normal World)", runtime.GOOS, runtime.GOARCH, runtime.Version())

	if imx6ul.Native {
		log.Printf("PL1 in Normal World is about to perform DCP key derivation")

		imx6ul.DCP.Init()

		// this fails after restrictions are in place (see trusted_os/tz.go)
		k, err := imx6ul.DCP.DeriveKey(make([]byte, 8), make([]byte, 16), -1)

		if err != nil {
			log.Printf("PL1 in Normal World World failed to use DCP (%v)", err)
		} else {
			log.Printf("PL1 in Normal World successfully used DCP (%x)", k)
		}

		// Uncomment to test memory protection, this will hang Normal World and
		// therefore everything.
		// mem.TestAccess("PL1 in Normal World")
	}

	// yield back to secure monitor
	log.Printf("PL1 in Normal World is about to yield back")
	exit()

	// this should be unreachable
	log.Printf("PL1 in Normal World says goodbye")
}

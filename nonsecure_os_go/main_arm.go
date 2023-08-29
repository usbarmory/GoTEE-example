// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"crypto/aes"
	"crypto/sha256"
	"log"
	"os"
	"runtime"
	_ "unsafe"

	"github.com/usbarmory/tamago/dma"
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
}

//go:linkname printk runtime.printk
func printk(c byte) {
	printSecure(c)
}

func init() {
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout)

	if !imx6ul.Native {
		return
	}

	switch imx6ul.Model() {
	case "i.MX6UL":
		imx6ul.SetARMFreq(imx6ul.Freq528)
		imx6ul.CAAM.DeriveKeyMemory = dma.Default()
	case "i.MX6ULL", "i.MX6ULZ":
		imx6ul.SetARMFreq(imx6ul.FreqMax)
		imx6ul.DCP.Init()
	}
}

func main() {
	log.Printf("%s/%s (%s) • system/supervisor (Non-secure)", runtime.GOOS, runtime.GOARCH, runtime.Version())

	if imx6ul.Native {
		var err error
		var k []byte

		log.Printf("supervisor is about to perform hardware key derivation")

		switch {
		case imx6ul.CAAM != nil:
			// derived key differs in non-secure
			k = make([]byte, sha256.Size)
			err = imx6ul.CAAM.DeriveKey(make([]byte, sha256.Size), k)
		case imx6ul.DCP != nil:
			// this fails after restrictions are in place (see trusted_os/tz.go)
			k, err = imx6ul.DCP.DeriveKey(make([]byte, aes.BlockSize), make([]byte, aes.BlockSize), -1)
		}

		if err != nil {
			log.Printf("supervisor failed to derive key (%v)", err)
		} else {
			log.Printf("supervisor successfully derived key (%x)", k)
		}

		// Uncomment to test memory protection, this will hang NS
		// context and therefore everything.
		// mem.TestAccess("Non-secure OS")
	}

	// yield back to secure monitor
	log.Printf("supervisor is about to yield back")
	exit()

	// this should be unreachable
	log.Printf("supervisor says goodbye")
}

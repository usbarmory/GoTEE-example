// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"log"
	"os"
	"runtime"
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/nxp/imx8mp"

	"github.com/usbarmory/GoTEE-example/mem"
)

//go:linkname ramStart runtime/goos.RamStart
var ramStart uint64 = mem.NonSecureStart

//go:linkname ramSize runtime/goos.RamSize
var ramSize uint64 = mem.NonSecureSize

//go:linkname hwinit runtime/goos.Hwinit1
func hwinit() {
	imx8mp.Init()
}

//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	printSecure(c)
}

func init() {
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout)
}

func main() {
	log.Printf("%s/%s (%s) • system/supervisor (Non-secure)", runtime.GOOS, runtime.GOARCH, runtime.Version())

	// uncomment to test memory protection
	//mem.TestAccess("Non-secure OS")

	// yield back to secure monitor
	log.Printf("supervisor is about to yield back")
	exit()

	// this should be unreachable
	log.Printf("supervisor says goodbye")
}

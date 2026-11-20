// Copyright (c) The GoTEE authors. All Rights Reserved.
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
	_ "unsafe"

	"github.com/usbarmory/tamago/board/nxp/imx8mpevk"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/nxp/imx8mp"

	"github.com/usbarmory/GoTEE-example/internal/semihosting"
	"github.com/usbarmory/GoTEE-example/mem"
	"github.com/usbarmory/GoTEE-example/trusted_os_imx8mpevk/cmd"
	"github.com/usbarmory/GoTEE-example/trusted_os_imx8mpevk/internal"
	"github.com/usbarmory/GoTEE-example/util"
)

// This example embeds the Trusted Applet and Main OS ELF binaries within the
// Trusted OS executable, using Go embed package.

//go:embed assets/trusted_applet.elf
var taELF []byte

//go:embed assets/nonsecure_os_go.elf
var osELF []byte

//go:linkname ramStart runtime/goos.RamStart
var ramStart uint64 = mem.SecureStart

//go:linkname ramSize runtime/goos.RamSize
var ramSize uint64 = mem.SecureSize

func init() {
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout)

	mem.Init()
	dma.Init(mem.SecureDMAStart, mem.SecureDMASize)

	cmd.Banner = fmt.Sprintf("%s/%s (%s) • TEE Security Monitor (Secure:%v EL%d)",
		runtime.GOOS, runtime.GOARCH, runtime.Version(),
		imx8mp.ARM64.Secure(), imx8mp.ARM64.CurrentEL())

	gotee.TA = taELF
	//gotee.OS = osELF (TODO)
}

func main() {
	gotee.Console = util.NewScreenConsole()
	cmd.SerialConsole(imx8mpevk.UART1)

	log.Printf("SM says goodbye")
	semihosting.Exit()
}

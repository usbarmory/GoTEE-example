// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
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
	"time"
	_ "unsafe"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"

	"github.com/usbarmory/imx-usbnet"

	"github.com/usbarmory/GoTEE-example/internal/semihosting"
	"github.com/usbarmory/GoTEE-example/mem"
	"github.com/usbarmory/GoTEE-example/util"

	"github.com/usbarmory/GoTEE-example/trusted_os_usbarmory/cmd"
	"github.com/usbarmory/GoTEE-example/trusted_os_usbarmory/internal"
)

const (
	sshPort = 22
	IP      = "10.0.0.1"
	MAC     = "1a:55:89:a2:69:41"
	hostMAC = "1a:55:89:a2:69:42"
)

// This example embeds the Trusted Applet and Main OS ELF binaries within the
// Trusted OS executable, using Go embed package.
//
// The loading strategy is up to implementers, on the NXP i.MX6 the armory-boot
// bootloader primitives can be used to create a bootable Trusted OS with
// authenticated disk loading of applets and kernels, see loadLinux() and:
//   https://pkg.go.dev/github.com/usbarmory/armory-boot

//go:embed assets/trusted_applet.elf
var taELF []byte

//go:embed assets/nonsecure_os_go.elf
var osELF []byte

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = mem.SecureStart

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = mem.SecureSize

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

	cmd.Banner = fmt.Sprintf("%s/%s (%s) • TEE security monitor (Secure World system/monitor)", runtime.GOOS, runtime.GOARCH, runtime.Version())

	gotee.TA = taELF
	gotee.OS = osELF
}

func serialConsole() {
	gotee.Console = util.NewScreenConsole()
	cmd.SerialConsole(usbarmory.UART2)

	log.Printf("SM says goodbye")
	semihosting.Exit()
}

func main() {
	if !imx6ul.Native {
		serialConsole()
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

	gotee.Console = &util.Console{
		Handler:  cmd.Handler,
		Listener: listener,
	}

	if err = gotee.Console.Start(); err != nil {
		log.Fatalf("SM could not initialize SSH server, %v", err)
	}

	usbarmory.USB1.Init()
	usbarmory.USB1.DeviceMode()
	usbarmory.USB1.Reset()

	// never returns
	usbarmory.USB1.Start(iface.NIC.Device)
}

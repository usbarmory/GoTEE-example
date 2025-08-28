// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gotee

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/bits"
	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/usdhc"

	"github.com/usbarmory/GoTEE/monitor"

	"github.com/usbarmory/GoTEE-example/mem"
	"github.com/usbarmory/GoTEE-example/util"

	"github.com/usbarmory/armory-boot/config"
	"github.com/usbarmory/armory-boot/disk"
	"github.com/usbarmory/armory-boot/exec"
)

// bootConfLinux is the path to the armory-boot configuration file for loading a
// Linux kernel as Non-secure OS.
const bootConfLinux = "/boot/armory-boot-nonsecure.conf"

var (
	TA []byte
	OS []byte
)

func configureMMU(region *dma.Region, alias uint32) {
	start := uint32(region.Start())
	end := uint32(region.End())

	imx6ul.ARM.ConfigureMMU(start, end, alias, arm.MemoryRegion|arm.TTE_AP_011<<10)
}

// loadApplet loads a TamaGo unikernel as trusted applet.
func loadApplet(lockstep bool) (ta *monitor.ExecCtx, err error) {
	image := &exec.ELFImage{
		Region: mem.AppletRegion,
		ELF:    TA,
	}

	alias := uint32(mem.AppletPhysicalStart)

	switch {
	case lockstep:
		log.Printf("SM loading applet in lockstep shadow memory")
		configureMMU(image.Region, mem.AppletShadowStart)

		if err = image.Load(); err != nil {
			return
		}
	case imx6ul.Native && imx6ul.BEE != nil && mem.BEE:
		log.Printf("SM loading applet in BEE encrypted memory")
		alias = 0
	}

	configureMMU(image.Region, alias)

	if err = image.Load(); err != nil {
		return
	}

	if ta, err = monitor.Load(image.Entry(), image.Region, true); err != nil {
		return nil, fmt.Errorf("SM could not load applet, %v", err)
	}

	log.Printf("SM loaded applet addr:%#x entry:%#x size:%d", ta.Memory.Start(), ta.R15, len(TA))

	// set applet as ELF debugging target
	util.SetDebugTarget(image.ELF)

	// register example RPC receiver
	ta.Server.Register(&RPC{})

	// set stack pointer to the end of available memory
	ta.R13 = uint32(ta.Memory.End())

	// override default handler to improve logging
	ta.Handler = goHandler

	if lockstep {
		ta.Shadow = ta.Clone()

		ta.MMU = func() {
			configureMMU(ta.Memory, mem.AppletPhysicalStart)
		}

		ta.Shadow.MMU = func() {
			configureMMU(ta.Memory, mem.AppletShadowStart)
		}
	}

	return
}

// loadNormalWorld loads a TamaGo unikernel as Normal World OS.
func loadNormalWorld(lock bool) (os *monitor.ExecCtx, err error) {
	image := &exec.ELFImage{
		Region: mem.NonSecureRegion,
		ELF:    OS,
	}

	if err = image.Load(); err != nil {
		return
	}

	if os, err = monitor.Load(image.Entry(), image.Region, false); err != nil {
		return nil, fmt.Errorf("SM could not load kernel, %v", err)
	}

	log.Printf("SM loaded kernel addr:%#x entry:%#x size:%d", os.Memory.Start(), os.R15, len(OS))

	if err = configureTrustZone(lock, false); err != nil {
		return nil, fmt.Errorf("SM could not configure TrustZone, %v", err)
	}

	// override default handler to handle exceptions and improve logging
	os.Handler = goHandler

	return
}

// loadLinux loads a Linux kernel as Normal World OS, the kernel configuration
// is read from an armory-boot configuration file on the given device ("eMMC"
// or "uSD").
func loadLinux(device string) (os *monitor.ExecCtx, err error) {
	var id int
	var card *usdhc.USDHC

	switch device {
	case "uSD":
		id = 10
		card = usbarmory.SD
	case "eMMC":
		id = 11
		card = usbarmory.MMC
	default:
		return nil, errors.New("invalid device")
	}

	// Set the device USDHC controller as Secure master to grant access to
	// the Trusted OS DMA region.
	if err = imx6ul.CSU.SetAccess(id, true, false); err != nil {
		return
	}

	part, err := disk.Detect(card, "")

	if err != nil {
		return
	}

	conf, err := config.Load(part, bootConfLinux, "", "")

	if err != nil {
		return
	}

	log.Printf("\n%s", conf.JSON)

	image := &exec.LinuxImage{
		Region:               mem.NonSecureRegion,
		Kernel:               conf.Kernel(),
		DeviceTreeBlob:       conf.DeviceTreeBlob(),
		InitialRamDisk:       conf.InitialRamDisk(),
		KernelOffset:         0x00800000,
		DeviceTreeBlobOffset: 0x07000000,
		InitialRamDiskOffset: 0x08000000,
		CmdLine:              conf.CmdLine,
	}

	if err = image.Load(); err != nil {
		return
	}

	if os, err = monitor.Load(image.Entry(), image.Region, false); err != nil {
		return nil, fmt.Errorf("SM could not load kernel, %v", err)
	}

	log.Printf("SM loaded kernel addr:%#x size:%d entry:%#x", os.Memory.Start(), len(image.Kernel), os.R15)

	if err = configureTrustZone(true, true); err != nil {
		return nil, fmt.Errorf("SM could not configure TrustZone, %v", err)
	}

	if err = grantPeripheralAccess(); err != nil {
		return nil, fmt.Errorf("SM could not configure TrustZone peripheral access, %v", err)
	}

	os.R0 = 0
	os.R2 = uint32(image.DTB())
	os.SPSR = arm.SVC_MODE

	// enable FIQ to receive TrustZone Watchdog IRQ
	bits.Clear(&os.SPSR, 6)

	// override default handler to service TrustZone Watchdog
	os.Handler = linuxHandler

	return
}

func run(ctx *monitor.ExecCtx, wg *sync.WaitGroup) {
	mode := arm.ModeName(int(ctx.SPSR) & 0x1f)
	ns := ctx.NonSecure()

	log.Printf("SM starting mode:%s sp:%#.8x pc:%#.8x ns:%v", mode, ctx.R13, ctx.R15, ns)

	err := ctx.Run()

	if wg != nil {
		wg.Done()
	}

	log.Printf("SM stopped mode:%s sp:%#.8x lr:%#.8x pc:%#.8x ns:%v err:%v %s", mode, ctx.R13, ctx.R14, ctx.R15, ns, err, ctx)

	if err != nil {
		if ctx.Shadow != nil {
			log.Printf("shadow context: %s", ctx.Shadow)
		}

		pcLine, _ := util.PCToLine(uint64(ctx.R15))
		lrLine, _ := util.PCToLine(uint64(ctx.R14))

		if pcLine != "" || lrLine != "" {
			log.Printf("stack trace:\n  %s\n  %s", pcLine, lrLine)
		}
	}
}

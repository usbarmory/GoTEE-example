// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	_ "embed"
	"fmt"
	"log"
	"sync"

	"github.com/f-secure-foundry/tamago/arm"

	"github.com/f-secure-foundry/GoTEE/monitor"
	"github.com/f-secure-foundry/GoTEE/syscall"

	"github.com/f-secure-foundry/GoTEE-example/mem"

	"github.com/f-secure-foundry/armory-boot/config"
	"github.com/f-secure-foundry/armory-boot/disk"
	"github.com/f-secure-foundry/armory-boot/exec"
)

// This example embeds the Trusted Applet and Main OS ELF binaries within the
// Trusted OS executable, using Go embed package.
//
// The loading strategy is up to implementers, on the NXP i.MX6 the armory-boot
// bootloader primitives can be used to create a bootable Trusted OS with
// authenticated disk loading of applets and kernels, see:
//   https://pkg.go.dev/github.com/f-secure-foundry/armory-boot

//go:embed assets/trusted_applet.elf
var taELF []byte

//go:embed assets/nonsecure_os_go.elf
var osELF []byte

// logHandler allows to override the GoTEE default handler and avoid
// interleaved logs, as the supervisor and applet contexts are logging
// simultaneously.
func logHandler(ctx *monitor.ExecCtx) (err error) {
	defaultHandler := monitor.SecureHandler

	if ctx.NonSecure() {
		defaultHandler = monitor.NonSecureHandler
	}

	if ctx.R0 == syscall.SYS_WRITE {
		if ssh != nil {
			bufferedTermLog(byte(ctx.R1), ctx.NonSecure(), ssh.Term)
		} else {
			bufferedStdoutLog(byte(ctx.R1), ctx.NonSecure())
		}
	} else {
		err = defaultHandler(ctx)
	}

	return
}

// loadApplet loads a TamaGo unikernel as trusted applet.
func loadApplet() (ta *monitor.ExecCtx, err error) {
	image := &exec.ELFImage{
		Region: mem.AppletRegion,
		ELF:    taELF,
	}

	if err = image.Load(); err != nil {
		return
	}

	if ta, err = monitor.Load(image.Entry(), image.Region, true); err != nil {
		return nil, fmt.Errorf("PL1 could not load applet, %v", err)
	} else {
		log.Printf("PL1 loaded applet addr:%#x size:%d entry:%#x", ta.Memory.Start, len(taELF), ta.R15)
	}

	// register example RPC receiver
	ta.Server.Register(&RPC{})
	ta.Debug = true

	// set stack pointer to the end of applet memory
	ta.R13 = mem.AppletStart + mem.AppletSize

	// override default handler to improve logging
	ta.Handler = logHandler

	return
}

// loadNormalWorld loads a TamaGo unikernel as normal world OS.
func loadNormalWorld(lock bool) (os *monitor.ExecCtx, err error) {
	image := &exec.ELFImage{
		Region: mem.NonSecureRegion,
		ELF:    osELF,
	}

	if err = image.Load(); err != nil {
		return
	}

	if os, err = monitor.Load(image.Entry(), image.Region, false); err != nil {
		return nil, fmt.Errorf("PL1 could not load kernel, %v", err)
	} else {
		log.Printf("PL1 loaded kernel addr:%#x size:%d entry:%#x", os.Memory.Start, len(osELF), os.R15)
	}

	os.Debug = true

	if err = configureTrustZone(lock); err != nil {
		return nil, fmt.Errorf("PL1 could not configure TrustZone, %v", err)
	}

	// override default handler to improve logging
	os.Handler = logHandler

	return
}

// loadDebian loads a Linux distribution as normal world OS, the kernel
// configuration is read as an armory-boot configuration file from the given
// device ("eMMC" or "uSD"), ext4 partition offset and path.
func loadDebian(device string, start string, configPath string) (os *monitor.ExecCtx, err error) {
	part, err := disk.Detect(device, start)

	if err != nil {
		return
	}

	conf, err := config.Load(part, configPath, "", "")

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
		return nil, fmt.Errorf("PL1 could not load kernel, %v", err)
	} else {
		log.Printf("PL1 loaded kernel addr:%#x size:%d entry:%#x", os.Memory.Start, len(image.Kernel), os.R15)
	}

	os.Debug = true

	if err = configureTrustZone(true); err != nil {
		return nil, fmt.Errorf("PL1 could not configure TrustZone, %v", err)
	}

	// CPU register 0 must be 0
	os.R0 = 0
	// CPU register 1 not required for DTB boot
	// CPU register 2 must be the parameter list address
	os.R2 = image.DTB()

	return
}

func run(ctx *monitor.ExecCtx, wg *sync.WaitGroup) {
	mode := arm.ModeName(int(ctx.SPSR) & 0x1f)
	ns := ctx.NonSecure()

	log.Printf("PL1 starting mode:%s ns:%v sp:%#.8x pc:%#.8x", mode, ns, ctx.R13, ctx.R15)

	err := ctx.Run()

	if wg != nil {
		wg.Done()
	}

	log.Printf("PL1 stopped mode:%s ns:%v sp:%#.8x lr:%#.8x pc:%#.8x err:%v", mode, ns, ctx.R13, ctx.R14, ctx.R15, err)
}

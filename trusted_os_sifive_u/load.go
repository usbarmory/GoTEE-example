// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/usbarmory/GoTEE/monitor"
	"github.com/usbarmory/GoTEE/syscall"

	"github.com/usbarmory/GoTEE-example/mem"
	"github.com/usbarmory/GoTEE-example/util"

	"github.com/usbarmory/armory-boot/exec"
)

// This example embeds the Trusted Applet and Main OS ELF binaries within the
// Trusted OS executable, using Go embed package.

//go:embed assets/trusted_applet.elf
var taELF []byte

//go:embed assets/nonsecure_os_go.elf
var osELF []byte
// logHandler allows to override the GoTEE default handler and avoid
// interleaved logs, as the supervisor and applet contexts are logging
// simultaneously.
func logHandler(ctx *monitor.ExecCtx) (err error) {
	defaultHandler := monitor.SecureHandler

	if !ctx.Secure() {
		defaultHandler = monitor.NonSecureHandler
	}

	switch {
	case ctx.A0() == syscall.SYS_WRITE:
		util.BufferedStdoutLog(byte(ctx.A1()), ctx.Secure())
	case !ctx.Secure() && ctx.A0() == syscall.SYS_EXIT:
		if ctx.Debug {
			ctx.Print()
		}

		return errors.New("exit")
	default:
		return defaultHandler(ctx)
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
		return nil, fmt.Errorf("SM could not load applet, %v", err)
	}

	log.Printf("SM loaded applet addr:%#x size:%d entry:%#x", ta.Memory.Start, len(taELF), ta.PC)

	// TODO: move this at context switch
	if err = configureAppletPMP(true); err != nil {
		return nil, fmt.Errorf("SM could not configure applet PMP, %v", err)
	}

	// register example RPC receiver
	ta.Server.Register(&RPC{})

	// set stack pointer to the end of applet memory
	ta.X2 = mem.AppletStart + mem.AppletSize

	// override default handler to improve logging
	ta.Handler = logHandler
	ta.Debug = true

	return
}

// loadSupervisor loads a TamaGo unikernel as main OS.
func loadSupervisor(lock bool) (os *monitor.ExecCtx, err error) {
	image := &exec.ELFImage{
		Region: mem.NonSecureRegion,
		ELF:    osELF,
	}

	if err = image.Load(); err != nil {
		return
	}

	if os, err = monitor.Load(image.Entry(), image.Region, false); err != nil {
		return nil, fmt.Errorf("SM could not load kernel, %v", err)
	}

	log.Printf("SM loaded kernel addr:%#x size:%d entry:%#x", os.Memory.Start, len(osELF), os.PC)

	//if err = configureSupervisorPMP(lock); err != nil {
	//	return nil, fmt.Errorf("SM could not configure supervisor PMP, %v", err)
	//}

	// override default handler to improve logging
	os.Handler = logHandler
	os.Debug = true

	return
}

func run(ctx *monitor.ExecCtx, wg *sync.WaitGroup) {
	log.Printf("SM starting sp:%#.8x pc:%#.8x", ctx.X2, ctx.PC)

	err := ctx.Run()

	if wg != nil {
		wg.Done()
	}

	log.Printf("SM stopped sp:%#.8x ra:%#.8x pc:%#.8x err:%v", ctx.X2, ctx.X1, ctx.PC, err)
}

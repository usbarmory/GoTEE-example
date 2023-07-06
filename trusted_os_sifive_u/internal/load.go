// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gotee

import (
	"fmt"
	"log"
	"sync"

	"github.com/usbarmory/GoTEE/monitor"

	"github.com/usbarmory/GoTEE-example/mem"

	"github.com/usbarmory/armory-boot/exec"
)

var (
	TA []byte
	OS []byte
)

// loadApplet loads a TamaGo unikernel as trusted applet.
func loadApplet() (ta *monitor.ExecCtx, err error) {
	image := &exec.ELFImage{
		Region: mem.AppletRegion,
		ELF:    TA,
	}

	if err = image.Load(); err != nil {
		return
	}

	if ta, err = monitor.Load(image.Entry(), image.Region, true); err != nil {
		return nil, fmt.Errorf("SM could not load applet, %v", err)
	}

	log.Printf("SM loaded applet addr:%#x entry:%#x size:%d", ta.Memory.Start(), ta.PC, len(TA))

	// set memory protection function
	ta.PMP = configurePMP

	// register example RPC receiver
	ta.Server.Register(&RPC{})

	// set stack pointer to the end of available memory
	ta.X2 = uint64(ta.Memory.End())

	// override default handler to improve logging
	ta.Handler = goHandler
	ta.Debug = true

	return
}

// loadSupervisor loads a TamaGo unikernel as main OS.
func loadSupervisor() (os *monitor.ExecCtx, err error) {
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

	log.Printf("SM loaded kernel addr:%#x entry:%#x size:%d", os.Memory.Start(), os.PC, len(OS))

	// set memory protection function
	os.PMP = configurePMP

	// set stack pointer to the end of available memory
	os.X2 = uint64(os.Memory.End())

	// override default handler to support SBI and improve logging
	os.Handler = sbiHandler
	os.Debug = true

	return
}

func run(ctx *monitor.ExecCtx, wg *sync.WaitGroup) {
	log.Printf("SM starting sp:%#.8x pc:%#.8x secure:%v", ctx.X2, ctx.PC, ctx.Secure())

	err := ctx.Run()

	if wg != nil {
		wg.Done()
	}

	log.Printf("SM stopped sp:%#.8x ra:%#.8x pc:%#.8x err:%v", ctx.X2, ctx.X1, ctx.PC, err)

	if err != nil {
		pcLine, _ := PCToLine(TA, ctx.PC)
		lrLine, _ := PCToLine(TA, ctx.X1)

		log.Printf("\t%s", pcLine)
		log.Printf("\t%s", lrLine)
	}
}

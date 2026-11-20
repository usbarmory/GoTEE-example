// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gotee

import (
	"fmt"
	"log"
	"sync"

	"github.com/usbarmory/tamago/arm64"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/nxp/imx8mp"

	"github.com/usbarmory/GoTEE/monitor"

	"github.com/usbarmory/GoTEE-example/mem"
	"github.com/usbarmory/GoTEE-example/util"

	"github.com/usbarmory/armory-boot/exec"
)

var TA []byte

func configureMMU(region *dma.Region, alias uint64) {
	start := uint64(region.Start())
	end := uint64(region.End())
	flags := arm64.MemoryAttributes | arm64.TTE_BLOCK | arm64.TTE_AP_01<<arm64.TTE_AP

	imx8mp.ARM64.ConfigureMMU(start, end, alias, flags)
}

// loadApplet loads a TamaGo unikernel as trusted applet.
func loadApplet() (ta *monitor.ExecCtx, err error) {
	image := &exec.ELFImage{
		Region: mem.AppletRegion,
		ELF:    TA,
	}

	configureMMU(image.Region, 0)

	if err = image.Load(); err != nil {
		return
	}

	if ta, err = monitor.Load(image.Entry(), image.Region, true); err != nil {
		return nil, fmt.Errorf("SM could not load applet, %v", err)
	}

	log.Printf("SM loaded applet addr:%#x entry:%#x size:%d", ta.Memory.Start(), ta.PC, len(TA))

	// set applet as ELF debugging target
	util.SetDebugTarget(image.ELF)

	// register example RPC receiver
	ta.Server.Register(&RPC{})

	// set stack pointer to the end of available memory
	ta.R31 = uint64(ta.Memory.End())

	// override default handler to improve logging
	ta.Handler = goHandler

	return
}

func run(ctx *monitor.ExecCtx, wg *sync.WaitGroup) {
	log.Printf("SM starting sp:%#.8x pc:%#.8x", ctx.R31, ctx.PC)

	err := ctx.Run()

	if wg != nil {
		wg.Done()
	}

	log.Printf("SM stopped lr:%#.8x sp:%#.8x pc:%#.8x err:%v %s", ctx.R30, ctx.R31, ctx.PC, err, ctx)

	if err != nil {
		pcLine, _ := util.PCToLine(ctx.PC)
		lrLine, _ := util.PCToLine(ctx.R30)

		if pcLine != "" || lrLine != "" {
			log.Printf("stack trace:\n  %s\n  %s", pcLine, lrLine)
		}
	}
}

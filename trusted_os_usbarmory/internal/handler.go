// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gotee

import (
	"errors"
	"fmt"
	"log"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"

	"github.com/usbarmory/GoTEE/monitor"
	"github.com/usbarmory/GoTEE/syscall"

	"github.com/usbarmory/GoTEE-example/util"
)

var Console *util.Console

func goHandler(ctx *monitor.ExecCtx) (err error) {
	if ctx.ExceptionVector == arm.DATA_ABORT && ctx.NonSecure() {
		log.Printf("SM trapped Non-secure data abort pc:%#.8x", ctx.R15-8)
		return
	}

	if ctx.ExceptionVector != arm.SUPERVISOR {
		return fmt.Errorf("exception %x", ctx.ExceptionVector)
	}

	switch ctx.A0() {
	case syscall.SYS_WRITE:
		// Override write syscall to avoid interleaved logs and to log
		// simultaneously to remote terminal and serial console.
		if Console != nil {
			util.BufferedTermLog(byte(ctx.A1()), !ctx.NonSecure(), Console.Term)
		} else {
			util.BufferedStdoutLog(byte(ctx.A1()), !ctx.NonSecure())
		}
	case syscall.SYS_EXIT:
		// support exit syscall on both security states
		ctx.Stop()
	default:
		if ctx.NonSecure() {
			ctx.Print()
			return errors.New("unexpected monitor call")
		} else {
			return monitor.SecureHandler(ctx)
		}
	}

	return
}

func linuxHandler(ctx *monitor.ExecCtx) (err error) {
	if !ctx.NonSecure() {
		return errors.New("unexpected processor mode")
	}

	if ctx.ExceptionVector == arm.FIQ {
		irq, end := imx6ul.GIC.GetInterrupt(true)

		if irq == imx6ul.TZ_WDOG.IRQ {
			log.Printf("SM servicing TrustZone Watchdog")
			imx6ul.TZ_WDOG.Service(watchdogTimeout)
		}

		if end != nil {
			end <- true
		}

		return
	}

	if ctx.ExceptionVector != arm.SUPERVISOR {
		return fmt.Errorf("exception %x", ctx.ExceptionVector)
	}

	return monitor.NonSecureHandler(ctx)
}

// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	_ "embed"
	"errors"
	"log"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"

	"github.com/usbarmory/GoTEE/monitor"
	"github.com/usbarmory/GoTEE/syscall"

	"github.com/usbarmory/GoTEE-example/util"
)

// logHandler allows to override the GoTEE default handler and avoid
// interleaved logs, as the supervisor and applet contexts are logging
// simultaneously.
func logHandler(ctx *monitor.ExecCtx) (err error) {
	switch {
	case ctx.A0() == syscall.SYS_WRITE:
		if ssh != nil {
			util.BufferedTermLog(byte(ctx.A1()), !ctx.NonSecure(), ssh.Term)
		} else {
			util.BufferedStdoutLog(byte(ctx.A1()), !ctx.NonSecure())
		}
	case ctx.NonSecure() && ctx.A0() == syscall.SYS_EXIT:
		if ctx.Debug {
			ctx.Print()
		}

		return errors.New("exit")
	default:
		if !ctx.NonSecure() {
			return monitor.SecureHandler(ctx)
		}
	}

	return
}

// linuxHandler services the TrustZone Watchdog
func linuxHandler(ctx *monitor.ExecCtx) (err error) {
	if !ctx.NonSecure() {
		panic("unexpected processor mode")
	}

	if ctx.ExceptionVector == arm.FIQ && imx6ul.ARM.GetInterrupt() == imx6ul.TZ_WDOG.IRQ {
		log.Printf("SM servicing TrustZone Watchdog")
		imx6ul.TZ_WDOG.Service(watchdogTimeout)

		// PC must be adjusted when returning from FIQ exceptions
		// (Table 11-3, ARM® Cortex™ -A Series Programmer’s Guide).
		ctx.R15 -= 4

		return
	}

	return monitor.NonSecureHandler(ctx)
}

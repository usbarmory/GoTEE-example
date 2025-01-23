// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gotee

import (
	"github.com/usbarmory/GoTEE/monitor"
	"github.com/usbarmory/GoTEE/sbi"
	"github.com/usbarmory/GoTEE/syscall"

	"github.com/usbarmory/GoTEE-example/util"
)

var Console *util.Console

func goHandler(ctx *monitor.ExecCtx) (err error) {
	defaultHandler := monitor.SecureHandler

	if !ctx.Secure() {
		defaultHandler = monitor.NonSecureHandler
	}

	switch {
	case ctx.A0() == syscall.SYS_WRITE:
		// Override write syscall to avoid interleaved logs and to log
		// simultaneously to remote terminal and serial console.
		if Console.Term != nil {
			util.BufferedTermLog(byte(ctx.A1()), ctx.Secure(), Console.Term)
		} else {
			util.BufferedStdoutLog(byte(ctx.A1()), ctx.Secure())
		}
	case !ctx.Secure() && ctx.A0() == syscall.SYS_EXIT:
		ctx.Stop()
	default:
		return defaultHandler(ctx)
	}

	return
}

func sbiHandler(ctx *monitor.ExecCtx) (err error) {
	// SBI v0.2 or higher calls are treated separately from GoTEE calls
	if ctx.X17 != 0 {
		return sbi.Handler(ctx)
	} else {
		return goHandler(ctx)
	}
}

// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gotee

import (
	"fmt"

	"github.com/usbarmory/tamago/arm64"

	"github.com/usbarmory/GoTEE/monitor"
	"github.com/usbarmory/GoTEE/syscall"

	"github.com/usbarmory/GoTEE-example/util"
)

var Console *util.Console

func goHandler(ctx *monitor.ExecCtx) (err error) {
	ec := (ctx.ESR >> arm64.ESR_EC) & 0x3f

	if ec != arm64.EC_SVC64 {
		// handle Go osInit ($GOROOT/src/internal/cpu/cpu_arm64_tamago.go)
		if ctx.ESR != 0 && (ec == arm64.EC_UNKNOWN || ec == arm64.EC_SYSREG) && ctx.EmulateIDReg() {
			return
		}

		return fmt.Errorf("exception ec:%#x esr:%#x", ec, ctx.ESR)
	}

	switch ctx.A0() {
	case syscall.SYS_WRITE:
		// Override write syscall to avoid interleaved logs and to log
		// simultaneously to remote terminal and serial console.
		if Console.Term != nil {
			util.BufferedTermLog(byte(ctx.A1()), true, Console.Term)
		} else {
			util.BufferedStdoutLog(byte(ctx.A1()), true)
		}
	case syscall.SYS_EXIT:
		// support exit syscall on both security states
		ctx.Stop()
	default:
		return monitor.SecureHandler(ctx)
	}

	return
}

// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/usbarmory/tamago/riscv"
	"github.com/usbarmory/tamago/soc/sifive/fu540"

	"github.com/usbarmory/GoTEE/monitor"

	"github.com/usbarmory/GoTEE-example/mem"
)

const (
	smStart = mem.SecureStart
	smEnd   = mem.SecureStart + mem.SecureSize + mem.SecureDMASize
)

func configurePMP(ctx *monitor.ExecCtx, i int) (err error) {
	// grant full peripheral access

	if err = fu540.RV64.WritePMP(i, 0x00000000, true, true, true, riscv.PMP_CFG_A_OFF, false); err != nil {
		return
	}

	if err = fu540.RV64.WritePMP(i+1, 0x80000000, true, true, true, riscv.PMP_CFG_A_TOR, false); err != nil {
		return
	}

	// protect Security Monitor

	if err = fu540.RV64.WritePMP(i+2, smStart, true, true, true, riscv.PMP_CFG_A_OFF, false); err != nil {
		return
	}

	if err = fu540.RV64.WritePMP(i+3, smEnd, false, false, false, riscv.PMP_CFG_A_TOR, false); err != nil {
		return
	}

	// TODO: IOPMP

	return
}

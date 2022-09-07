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

func protectSM(i int) (int, error) {
	if err := fu540.RV64.WritePMP(i, smStart, true, true, true, riscv.PMP_CFG_A_OFF, false); err != nil {
		return -1, err
	}
	i += 1

	if err := fu540.RV64.WritePMP(i, smEnd, false, false, false, riscv.PMP_CFG_A_TOR, false); err != nil {
		return -1, err
	}
	i += 1

	return i, nil
}

func configurePMP(ctx *monitor.ExecCtx) (err error) {
	var i int

	ctxStart := ctx.Memory.Start()
	ctxEnd := ctx.Memory.End()

	// grant full peripheral access
	if err = fu540.RV64.WritePMP(i, 0x80000000, true, true, true, riscv.PMP_CFG_A_TOR, false); err != nil {
		return
	}
	i += 1

	// Due to PMP priority evaluation the restriction for Secure Monitor
	// region is prioritized based on its location respective to the
	// execution context memory.
	switch {
	case ctxStart >= smEnd:
		if i, err = protectSM(i); err != nil {
			return
		}
	case ctxEnd <= smStart:
		defer func() {
			_, err = protectSM(i)
		}()
	default:
		panic("invalid memory layout")
	}

	// grant execution context memory region

	if err = fu540.RV64.WritePMP(i, uint64(ctxStart), true, true, true, riscv.PMP_CFG_A_OFF, false); err != nil {
		return
	}
	i += 1

	if err = fu540.RV64.WritePMP(i, uint64(ctxEnd), true, true, true, riscv.PMP_CFG_A_TOR, false); err != nil {
		return
	}
	i += 1

	// TODO: IOPMP

	return
}

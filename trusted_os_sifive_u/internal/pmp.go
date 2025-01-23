// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gotee

import (
	"github.com/usbarmory/tamago/riscv64"
	"github.com/usbarmory/tamago/soc/sifive/fu540"

	"github.com/usbarmory/GoTEE/monitor"

	"github.com/usbarmory/GoTEE-example/mem"
)

const (
	smStart = mem.SecureStart
	smEnd   = mem.SecureStart + mem.SecureSize + mem.SecureDMASize
)

func configurePMP(ctx *monitor.ExecCtx, i int) (err error) {
	// The main OS used in GoTEE-example, for the riscv64 architecture, is
	// a TamaGo unikernel which requires only PRCI, CLINT and UART0 access.
	//
	// On the FU540 the lack of IOPMP entails that only bus peripherals can
	// be given access through PMP, while bus controllers (e.g.  Ethernet)
	// must be exposed only through the Security Monitor API, and never
	// directly, for secure isolation.
	//
	// The access to PRCI, CLINT and UART0 can be granted without concerns
	// as they they cannot act as bus controllers.

	// grant PRCI and UART0 access

	if err = fu540.RV64.WritePMP(i, fu540.PRCI_BASE, false, false, false, riscv64.PMP_A_OFF, false); err != nil {
		return
	}
	i += 1

	if err = fu540.RV64.WritePMP(i, fu540.UART1_BASE, true, true, true, riscv64.PMP_A_TOR, false); err != nil {
		return
	}
	i += 1

	// grant CLINT access

	if err = fu540.RV64.WritePMP(i, fu540.CLINT_BASE, false, false, false, riscv64.PMP_A_OFF, false); err != nil {
		return
	}
	i += 1

	if err = fu540.RV64.WritePMP(i, fu540.CLINT_BASE+0x10000, true, true, true, riscv64.PMP_A_TOR, false); err != nil {
		return
	}
	i += 1

	// protect Security Monitor

	if err = fu540.RV64.WritePMP(i, smStart, false, false, false, riscv64.PMP_A_OFF, false); err != nil {
		return
	}
	i += 1

	if err = fu540.RV64.WritePMP(i, smEnd, false, false, false, riscv64.PMP_A_TOR, false); err != nil {
		return
	}

	return
}

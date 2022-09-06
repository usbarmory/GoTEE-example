// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/usbarmory/tamago/riscv"
	"github.com/usbarmory/tamago/soc/sifive/fu540"

	"github.com/usbarmory/GoTEE-example/mem"
)

func configureAppletPMP(lock bool) (err error) {
	if lock {
		// restrict Secure Monitor memory

		if err = fu540.RV64.WritePMP(0, mem.SecureStart, true, true, true, riscv.PMP_CFG_A_TOR, false); err != nil {
			return
		}

		if err = fu540.RV64.WritePMP(1, mem.SecureStart+mem.SecureSize+mem.SecureDMASize, false, false, false, riscv.PMP_CFG_A_TOR, false); err != nil {
			return
		}

		// grant Secure Monitor applet region

		if err = fu540.RV64.WritePMP(2, mem.AppletStart, true, true, true, riscv.PMP_CFG_A_TOR, false); err != nil {
			return
		}

		if err = fu540.RV64.WritePMP(3, mem.AppletStart+mem.AppletSize, true, true, true, riscv.PMP_CFG_A_TOR, false); err != nil {
			return
		}
	} else {
		// set default PMP region (entire memory space) to Supervisor/User mode access
		if err = fu540.RV64.WritePMP(0, (1 << 64) - 1, true, true, true, riscv.PMP_CFG_A_TOR, false); err != nil {
			return
		}
	}

	return
}

// TODO
//func configureSupervisorPMP(lock bool) (err error) {
//}

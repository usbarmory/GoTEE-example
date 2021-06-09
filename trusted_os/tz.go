// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/f-secure-foundry/tamago/soc/imx6"
	"github.com/f-secure-foundry/tamago/soc/imx6/csu"
	"github.com/f-secure-foundry/tamago/soc/imx6/tzasc"
)

func configureTrustZone(start uint32, size int, lock bool) (err error) {
	// grant NonSecure access to CP10 and CP11
	imx6.ARM.NonSecureAccessControl(1<<11 | 1<<10)

	if !imx6.Native {
		return
	}

	csu.Init()

	// grant NonSecure access to all peripherals
	for i := csu.CSL_MIN; i < csu.CSL_MAX; i++ {
		if err = csu.SetSecurityLevel(i, 0, csu.SEC_LEVEL_0, false); err != nil {
			return
		}

		if err = csu.SetSecurityLevel(i, 1, csu.SEC_LEVEL_0, false); err != nil {
			return
		}
	}

	// TZASC NonSecure World R/W access
	if err = tzasc.EnableRegion(1, start, size, (1<<tzasc.SP_NW_RD)|(1<<tzasc.SP_NW_WR)); err != nil {
		return
	}

	if !lock {
		return
	}

	// set all masters to NonSecure
	for i := csu.SA_MIN; i < csu.SA_MAX; i++ {
		if err = csu.SetAccess(i, false, false); err != nil {
			return
		}
	}

	// restrict ROMCP
	if err = csu.SetSecurityLevel(13, 0, csu.SEC_LEVEL_4, false); err != nil {
		return
	}

	// restrict TZASC
	if err = csu.SetSecurityLevel(16, 1, csu.SEC_LEVEL_4, false); err != nil {
		return
	}

	// restrict LEDs (GPIO4, IOMUXC)
	if err = csu.SetSecurityLevel(2, 1, csu.SEC_LEVEL_4, false); err != nil {
		return
	}

	if err = csu.SetSecurityLevel(6, 1, csu.SEC_LEVEL_4, false); err != nil {
		return
	}

	// restrict DCP
	if err = csu.SetSecurityLevel(34, 0, csu.SEC_LEVEL_4, false); err != nil {
		return
	}

	if err = csu.SetAccess(14, true, false); err != nil {
		return
	}

	if imx6.Native {
		// restrict USB
		if err = csu.SetSecurityLevel(4, 0, csu.SEC_LEVEL_4, false); err != nil {
			return
		}

		// set USB master as Secure
		if err = csu.SetAccess(4, true, false); err != nil {
			return
		}
	}

	return
}

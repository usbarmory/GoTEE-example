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

	"github.com/f-secure-foundry/GoTEE-example/mem"
)

func configureTrustZone(lock bool, usb bool, led bool) (err error) {
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

	// set default TZASC region (entire memory space) to NonSecure access
	if err = tzasc.EnableRegion(0, 0, 0, (1<<tzasc.SP_NW_RD)|(1<<tzasc.SP_NW_WR)); err != nil {
		return
	}

	if lock {
		// restrict Secure World memory
		if err = tzasc.EnableRegion(1, mem.SecureStart, mem.SecureSize + mem.SecureDMASize, (1<<tzasc.SP_SW_RD)|(1<<tzasc.SP_SW_WR)); err != nil {
			return
		}

		// restrict Secure World applet region
		if err = tzasc.EnableRegion(2, mem.AppletStart, mem.AppletSize, (1<<tzasc.SP_SW_RD)|(1<<tzasc.SP_SW_WR)); err != nil {
			return
		}
	} else {
		return
	}

	// set all controllers to NonSecure
	for i := csu.SA_MIN; i < csu.SA_MAX; i++ {
		if err = csu.SetAccess(i, false, false); err != nil {
			return
		}
	}

	// restrict access to ROMCP
	if err = csu.SetSecurityLevel(13, 0, csu.SEC_LEVEL_4, false); err != nil {
		return
	}

	// restrict access to TZASC
	if err = csu.SetSecurityLevel(16, 1, csu.SEC_LEVEL_4, false); err != nil {
		return
	}

	// restrict access to DCP
	if err = csu.SetSecurityLevel(34, 0, csu.SEC_LEVEL_4, false); err != nil {
		return
	}

	// set DCP as Secure
	if err = csu.SetAccess(14, true, false); err != nil {
		return
	}

	securityLevel := uint8(csu.SEC_LEVEL_0)
	secureAccess := false

	if usb {
		securityLevel = csu.SEC_LEVEL_4
		secureAccess = true
	}

	// USB
	if err = csu.SetSecurityLevel(8, 0, securityLevel, false); err != nil {
		return
	}

	// set USB controller as Secure
	if err = csu.SetAccess(4, secureAccess, false); err != nil {
		return
	}

	if led {
		securityLevel = csu.SEC_LEVEL_4
	} else {
		securityLevel = csu.SEC_LEVEL_0
	}

	// LEDs (GPIO4)
	if err = csu.SetSecurityLevel(2, 1, securityLevel, false); err != nil {
		return
	}

	// LEDs (IOMUXC)
	if err = csu.SetSecurityLevel(6, 1, securityLevel, false); err != nil {
		return
	}

	return
}

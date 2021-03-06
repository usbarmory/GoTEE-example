// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/usbarmory/tamago/soc/imx6"
	"github.com/usbarmory/tamago/soc/imx6/csu"
	"github.com/usbarmory/tamago/soc/imx6/imx6ul"
	"github.com/usbarmory/tamago/soc/imx6/tzasc"

	"github.com/usbarmory/GoTEE-example/mem"
)

func configureTrustZone(lock bool) (err error) {
	// grant NonSecure access to CP10 and CP11
	imx6.ARM.NonSecureAccessControl(1<<11 | 1<<10)

	if !imx6.Native {
		return
	}

	// grant NonSecure access to all peripherals
	for i := csu.CSL_MIN; i < csu.CSL_MAX; i++ {
		if err = imx6ul.CSU.SetSecurityLevel(i, 0, csu.SEC_LEVEL_0, false); err != nil {
			return
		}

		if err = imx6ul.CSU.SetSecurityLevel(i, 1, csu.SEC_LEVEL_0, false); err != nil {
			return
		}
	}

	// set default TZASC region (entire memory space) to NonSecure access
	if err = imx6ul.TZASC.EnableRegion(0, 0, 0, (1<<tzasc.SP_NW_RD)|(1<<tzasc.SP_NW_WR)); err != nil {
		return
	}

	if lock {
		// restrict Secure World memory
		if err = imx6ul.TZASC.EnableRegion(1, mem.SecureStart, mem.SecureSize+mem.SecureDMASize, (1<<tzasc.SP_SW_RD)|(1<<tzasc.SP_SW_WR)); err != nil {
			return
		}

		// restrict Secure World applet region
		if err = imx6ul.TZASC.EnableRegion(2, mem.AppletStart, mem.AppletSize, (1<<tzasc.SP_SW_RD)|(1<<tzasc.SP_SW_WR)); err != nil {
			return
		}
	} else {
		return
	}

	// set all controllers to NonSecure
	for i := csu.SA_MIN; i < csu.SA_MAX; i++ {
		if err = imx6ul.CSU.SetAccess(i, false, false); err != nil {
			return
		}
	}

	// restrict access to GPIO4 (used by LEDs)
	if err = imx6ul.CSU.SetSecurityLevel(2, 1, csu.SEC_LEVEL_4, false); err != nil {
		return
	}

	// restrict access to IOMUXC (used by LEDs)
	if err = imx6ul.CSU.SetSecurityLevel(6, 1, csu.SEC_LEVEL_4, false); err != nil {
		return
	}

	// restrict access to USB
	if err = imx6ul.CSU.SetSecurityLevel(8, 0, csu.SEC_LEVEL_4, false); err != nil {
		return
	}

	// set USB controller as Secure
	if err = imx6ul.CSU.SetAccess(4, true, false); err != nil {
		return
	}

	// restrict access to ROMCP
	if err = imx6ul.CSU.SetSecurityLevel(13, 0, csu.SEC_LEVEL_4, false); err != nil {
		return
	}

	// restrict access to TZASC
	if err = imx6ul.CSU.SetSecurityLevel(16, 1, csu.SEC_LEVEL_4, false); err != nil {
		return
	}

	// restrict access to DCP
	if err = imx6ul.CSU.SetSecurityLevel(34, 0, csu.SEC_LEVEL_4, false); err != nil {
		return
	}

	// set DCP as Secure
	if err = imx6ul.CSU.SetAccess(14, true, false); err != nil {
		return
	}

	return
}

func grantPeripheralAccess() (err error) {
	// allow access to GPIO4 (used by LEDs)
	if err = imx6ul.CSU.SetSecurityLevel(2, 1, csu.SEC_LEVEL_0, false); err != nil {
		return
	}

	// allow access to IOMUXC (used by LEDs)
	if err = imx6ul.CSU.SetSecurityLevel(6, 1, csu.SEC_LEVEL_0, false); err != nil {
		return
	}

	// allow access to USB
	if err = imx6ul.CSU.SetSecurityLevel(8, 0, csu.SEC_LEVEL_0, false); err != nil {
		return
	}

	// set USB controller as NonSecure
	if err = imx6ul.CSU.SetAccess(4, false, false); err != nil {
		return
	}

	// set USDHC1 (microSD) controller as NonSecure
	if err = imx6ul.CSU.SetAccess(10, false, false); err != nil {
		return
	}

	// set USDHC2 (eMMC) controller as NonSecure
	if err = imx6ul.CSU.SetAccess(11, false, false); err != nil {
		return
	}

	return
}

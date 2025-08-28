// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gotee

import (
	"github.com/usbarmory/tamago/arm/tzc380"
	"github.com/usbarmory/tamago/soc/nxp/csu"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"

	"github.com/usbarmory/GoTEE-example/mem"
)

func configureTrustZone(lock bool, wdog bool) (err error) {
	// grant NonSecure access to CP10 and CP11
	imx6ul.ARM.NonSecureAccessControl(1<<11 | 1<<10)

	if !imx6ul.Native {
		return
	}

	// grant NonSecure access to all peripherals
	for i := csu.CSL_MIN; i <= csu.CSL_MAX; i++ {
		if err = imx6ul.CSU.SetSecurityLevel(i, 0, csu.SEC_LEVEL_0, false); err != nil {
			return
		}

		if err = imx6ul.CSU.SetSecurityLevel(i, 1, csu.SEC_LEVEL_0, false); err != nil {
			return
		}
	}

	if imx6ul.CAAM != nil {
		// set CAAM as NonSecure
		imx6ul.CAAM.SetOwner(false)
	}

	// set default TZASC region (entire memory space) to NonSecure access
	if err = imx6ul.TZASC.EnableRegion(0, 0, 0, (1<<tzc380.SP_NW_RD)|(1<<tzc380.SP_NW_WR)); err != nil {
		return
	}

	// enable OCRAM TrustZone support
	if err = imx6ul.SetOCRAMProtection(imx6ul.OCRAM_START); err != nil {
		return
	}

	// set ARM debugging
	imx6ul.Debug(!lock)

	if lock {
		// restrict Secure World memory
		if err = imx6ul.TZASC.EnableRegion(1, mem.SecureStart, mem.SecureSize+mem.SecureDMASize+mem.AppletSize, (1<<tzc380.SP_SW_RD)|(1<<tzc380.SP_SW_WR)); err != nil {
			return
		}

		// restrict Secure World applet virtual region
		if err = imx6ul.TZASC.EnableRegion(3, mem.AppletVirtualStart, mem.AppletSize, (1<<tzc380.SP_SW_RD)|(1<<tzc380.SP_SW_WR)); err != nil {
			return
		}
	} else {
		return
	}

	// set all controllers to NonSecure
	for i := csu.SA_MIN; i <= csu.SA_MAX; i++ {
		if err = imx6ul.CSU.SetAccess(i, false, false); err != nil {
			return
		}
	}

	// restrict access to GPIO4 (used by LEDs)
	if err = imx6ul.CSU.SetSecurityLevel(2, 1, csu.SEC_LEVEL_4, false); err != nil {
		return
	}

	if wdog {
		// restrict access to WDOG2 (TZ WDOG)
		if err = imx6ul.CSU.SetSecurityLevel(5, 0, csu.SEC_LEVEL_4, false); err != nil {
			return
		}
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

	if imx6ul.DCP != nil {
		// restrict access to DCP
		if err = imx6ul.CSU.SetSecurityLevel(34, 0, csu.SEC_LEVEL_4, false); err != nil {
			return
		}

		// set DCP as Secure
		if err = imx6ul.CSU.SetAccess(14, true, false); err != nil {
			return
		}
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

func enableTrustZoneWatchdog() {
	// initialize interrupt controller, route all interrupts to NonSecure
	imx6ul.GIC.Init(false, true)

	// enable TrustZone Watchdog Secure interrupt
	imx6ul.GIC.EnableInterrupt(imx6ul.TZ_WDOG.IRQ, true)
	imx6ul.TZ_WDOG.EnableInterrupt(watchdogWarningInterval)

	// enable TrustZone Watchdog
	imx6ul.TZ_WDOG.EnableTimeout(watchdogTimeout)
}

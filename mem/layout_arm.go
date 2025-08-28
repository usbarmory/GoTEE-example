// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mem

import (
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/nxp/bee"
)

const (
	// Secure Monitor
	SecureStart = 0x90000000
	SecureSize  = 0x05f00000 // 95MB

	// Secure Monitor DMA (relocated to avoid conflicts with Main OS)
	SecureDMAStart = 0x95f00000
	SecureDMASize  = 0x00100000 // 1MB

	// Secure Monitor Applet (virtual)
	AppletVirtualStart = bee.AliasRegion0 // memory alias
	AppletSize         = 0x02000000       // 32MB

	// Secure Monitor Applet (physical)
	//
	// i.MX6ULL/i.MX6ULZ: primary and shadow areas used in soft lockstep.
	//          i.MX6UL : primary area AES encrypted w/ BEE, no lockstep.
	AppletPhysicalStart = 0x96000000
	AppletShadowStart   = 0x98000000

	// Main OS
	NonSecureStart = 0x80000000
	NonSecureSize  = 0x10000000 // 256MB
)

// BEE enables AES CTR encryption for the Applet RAM on i.MX6UL P/Ns
const BEE = true

const textStartWord = 0xe59a1008

var (
	AppletRegion    *dma.Region
	NonSecureRegion *dma.Region
)

func Init() {
	AppletRegion, _ = dma.NewRegion(AppletVirtualStart, AppletSize, false)
	AppletRegion.Reserve(AppletSize, 0)

	NonSecureRegion, _ = dma.NewRegion(NonSecureStart, NonSecureSize, false)
	NonSecureRegion.Reserve(NonSecureSize, 0)
}

// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
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
	SecureStart = 0x98000000
	SecureSize  = 0x03f00000 // 63MB

	// Secure Monitor DMA (relocated to avoid conflicts with Main OS)
	SecureDMAStart = 0x9bf00000
	SecureDMASize  = 0x00100000 // 1MB

	// Secure Monitor Applet
	AppletPhysicalStart = 0x9c000000       // encrypted w/ BEE on i.MX6UL
	AppletVirtualStart  = bee.AliasRegion0 // memory alias
	AppletSize          = 0x02000000       // 32MB

	// Main OS
	NonSecureStart = 0x80000000
	NonSecureSize  = 0x10000000 // 256MB
)

// BEE enables AES CTR encryption for the Applet RAM on i.MX6UL P/Ns
const BEE = true

const textStartWord = 0xe59a1008

var AppletRegion *dma.Region
var NonSecureRegion *dma.Region

func Init() {
	AppletRegion, _ = dma.NewRegion(AppletVirtualStart, AppletSize, false)
	AppletRegion.Reserve(AppletSize, 0)

	NonSecureRegion, _ = dma.NewRegion(NonSecureStart, NonSecureSize, false)
	NonSecureRegion.Reserve(NonSecureSize, 0)
}

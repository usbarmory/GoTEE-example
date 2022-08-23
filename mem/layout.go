// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mem

import (
	"github.com/usbarmory/tamago/dma"
)

const (
	// Secure Monitor
	SecureStart = 0x98000000
	SecureSize  = 0x03f00000 // 63MB

	// Secure Monitor DMA (relocated to avoid conflicts with Main OS)
	SecureDMAStart = 0x9bf00000
	SecureDMASize  = 0x00100000 // 1MB

	// Secure Monitor Applet
	AppletStart = 0x9c000000
	AppletSize  = 0x02000000 // 32MB

	// Main OS
	NonSecureStart = 0x80000000
	NonSecureSize  = 0x10000000 // 256MB
)

var AppletRegion *dma.Region
var NonSecureRegion *dma.Region

func init() {
	AppletRegion, _ = dma.NewRegion(AppletStart, AppletSize, false)
	AppletRegion.Reserve(AppletSize, 0)

	NonSecureRegion, _ = dma.NewRegion(NonSecureStart, NonSecureSize, false)
	NonSecureRegion.Reserve(NonSecureSize, 0)
}

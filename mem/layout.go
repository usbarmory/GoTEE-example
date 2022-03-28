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
	// Secure World OS
	SecureStart = 0x98000000
	SecureSize  = 0x03f00000 // 63MB

	// Secure World DMA (relocated to avoid conflicts with NonSecure world)
	SecureDMAStart = 0x9bf00000
	SecureDMASize  = 0x00100000 // 1MB

	// Secure World Applet
	AppletStart = 0x9c000000
	AppletSize  = 0x02000000 // 32MB

	// NonSecure World OS
	NonSecureStart = 0x80000000
	NonSecureSize  = 0x10000000 // 256MB
)

var AppletRegion *dma.Region
var NonSecureRegion *dma.Region

func init() {
	AppletRegion = &dma.Region{
		Start: AppletStart,
		Size: AppletSize,
	}

	AppletRegion.Init()
	AppletRegion.Reserve(AppletSize, 0)

	NonSecureRegion = &dma.Region{
		Start: NonSecureStart,
		Size: NonSecureSize,
	}

	NonSecureRegion.Init()
	NonSecureRegion.Reserve(NonSecureSize, 0)
}

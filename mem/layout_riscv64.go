// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mem

import (
	"github.com/usbarmory/tamago/dma"
)

const (
	// Secure Monitor
	SecureStart = 0x90000000
	SecureSize  = 0x07f00000 // 127MB

	// Secure Monitor DMA (relocated to avoid conflicts with Main OS)
	SecureDMAStart = 0x97f00000
	SecureDMASize  = 0x00100000 // 1MB

	// Secure Monitor Applet
	AppletStart = 0x98000000
	AppletSize  = 0x04000000 // 64MB

	// Main OS
	NonSecureStart = 0x80000000
	NonSecureSize  = 0x10000000 // 256MB
)

const textStartWord = 0x010db303

var (
	AppletRegion    *dma.Region
	NonSecureRegion *dma.Region
)

func Init() {
	AppletRegion, _ = dma.NewRegion(AppletStart, AppletSize, false)
	AppletRegion.Reserve(AppletSize, 0)

	NonSecureRegion, _ = dma.NewRegion(NonSecureStart, NonSecureSize, false)
	NonSecureRegion.Reserve(NonSecureSize, 0)
}

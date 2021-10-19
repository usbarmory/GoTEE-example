// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mem

const (
	// Secure World OS
	SecureStart = 0x90000000
	SecureSize  = 0x03f00000 // 63MB

	// Secure World DMA (relocated to avoid conflicts with NonSecure world)
	SecureDMAStart = 0x93f00000
	SecureDMASize  = 0x00100000 // 1MB

	// Secure World Applet
	AppletStart = 0x94000000
	AppletSize  = 0x02000000 // 32MB

	// NonSecure World OS
	NonSecureStart = 0x80000000
	NonSecureSize  = 0x10000000 // 256MB
)

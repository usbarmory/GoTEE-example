// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mem

// This example memory layout allocates 32MB for each execution context.
const (
	// Secure World OS
	SecureStart = 0x80000000
	SecureSize  = 0x01f00000

	// Secure World DMA (relocated to avoid conflicts with NonSecure world)
	SecureDMAStart = 0x81f00000
	SecureDMASize  = 0x00100000

	// Secure World Applet
	AppletStart = 0x82000000
	AppletSize  = 0x02000000

	// NonSecure World OS
	NonSecureStart = 0x84000000
	NonSecureSize  = 0x02000000
)

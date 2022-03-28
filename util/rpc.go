// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package util

// LEDStatus represents an RPC LED state request.
type LEDStatus struct {
	// Name is the LED name
	Name string
	// On is the LED state
	On bool
}

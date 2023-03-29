// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package gotee

import (
	"errors"

	"github.com/usbarmory/GoTEE-example/util"
)

// RPC represents an example receiver for user mode <--> system RPC over system
// calls.
type RPC struct{}

// Echo returns a response with the input string.
func (r *RPC) Echo(in string, out *string) error {
	*out = in
	return nil
}

// LED receives a LED state request.
func (r *RPC) LED(led util.LEDStatus, _ *bool) error {
	switch led.Name {
	case "white", "White", "WHITE":
		return errors.New("LED is secure only")
	case "blue", "Blue", "BLUE":
		// unimplemented as the target is virtual for now
	default:
		return errors.New("invalid LED")
	}

	return nil
}

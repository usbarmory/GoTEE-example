// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	_ "unsafe"

	"github.com/usbarmory/GoTEE-example/mem"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint64 = mem.AppletStart

//go:linkname ramSize runtime.ramSize
var ramSize uint64 = mem.AppletSize

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint64 = 0x100

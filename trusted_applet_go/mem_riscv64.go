// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	_ "unsafe"

	"github.com/usbarmory/GoTEE-example/mem"
)

//go:linkname ramStart runtime/goos.RamStart
var ramStart uint64 = mem.AppletStart

//go:linkname ramSize runtime/goos.RamSize
var ramSize uint64 = mem.AppletSize

//go:linkname ramStackOffset runtime/goos.RamStackOffset
var ramStackOffset uint64 = 0x100

// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/usbarmory/GoTEE/syscall"
)

const (
	SYS_WRITE = syscall.SYS_WRITE
	SYS_EXIT  = syscall.SYS_EXIT
)

// defined in api_*.s
func printSecure(byte)
func exit()

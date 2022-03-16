// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
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

// defined in api.s
func printSecure(byte)
func exit()

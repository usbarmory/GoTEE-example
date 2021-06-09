// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/f-secure-foundry/GoTEE/syscall"
)

const (
	SYS_WRITE = syscall.SYS_WRITE
)

// defined in api.s
func printSecure(byte)

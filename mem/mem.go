// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package mem

import (
	"log"
	"sync/atomic"
	"unsafe"
)

// TestAccess attempts to read one 32-bit word from Secure World memory.
func TestAccess(tag string) {
	pl1TextStart := SecureStart + uint32(0x10000)
	mem := (*uint32)(unsafe.Pointer(uintptr(pl1TextStart)))

	log.Printf("%s is about to read PL1 Secure World memory at %#x", tag, pl1TextStart)
	val := atomic.LoadUint32(mem)

	res := "success - *insecure configuration*"

	if val != 0xe59a1008 {
		res = "fail (expected, but you should never see this)"
	}

	log.Printf("%s read PL1 Secure World memory %#x: %#x (%s)", tag, pl1TextStart, val, res)
}

// TestDataAbort abort attempts a write to unallocated memory.
func TestDataAbort(tag string) {
	var p *byte

	log.Printf("%s is about to trigger a data abort")
	*p = 0xab
}

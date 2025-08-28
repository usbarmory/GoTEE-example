// Copyright (c) The GoTEE authors. All Rights Reserved.
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
	addr := SecureStart + uint32(0x10000)
	mem := (*uint32)(unsafe.Pointer(uintptr(addr)))

	log.Printf("%s is about to read secure memory at %#x", tag, addr)
	val := atomic.LoadUint32(mem)

	res := "success - *insecure configuration*"

	if val != textStartWord {
		res = "fail (expected, but you should never see this)"
	}

	log.Printf("%s read secure memory %#x: %#x (%s)", tag, addr, val, res)
}

// TestDataAbort abort attempts a write to unallocated memory.
func TestDataAbort(tag string) {
	var p *byte

	log.Printf("%s is about to trigger a data abort", tag)
	*p = 0xab
}

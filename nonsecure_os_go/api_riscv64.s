// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"

// A7 must be set to 0 to avoid interference with SBI

// func printSecure(c byte)
TEXT ·printSecure(SB),$0-1
	MOV	$const_SYS_WRITE, A0
	MOV	c+0(FP), A1

	MOV	$0, A7
	ECALL

	RET

// func exit()
TEXT ·exit(SB),$0
	MOV	$const_SYS_EXIT, A0

	MOV	$0, A7
	ECALL

	RET

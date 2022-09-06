// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"

// func printSecure(c byte)
TEXT ·printSecure(SB),$0-1
	MOV	$const_SYS_WRITE, A0
	MOV	c+0(FP), A1

	ECALL

	RET

// func exit()
TEXT ·exit(SB),$0
	MOV	$const_SYS_EXIT, A0

	ECALL

	RET

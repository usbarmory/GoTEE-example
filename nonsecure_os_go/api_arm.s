// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"

// func printSecure(c byte)
TEXT ·printSecure(SB),$0-1
	MOVW	$const_SYS_WRITE, R0
	MOVB	c+0(FP), R1

	WORD	$0xe1600070 // smc 0

	RET

// func exit()
TEXT ·exit(SB),$0
	MOVW	$const_SYS_EXIT, R0

	WORD	$0xe1600070 // smc 0

	RET

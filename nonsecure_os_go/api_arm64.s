// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"

// func printSecure(c byte)
TEXT ·printSecure(SB),$0-1
	MOVD	$const_SYS_WRITE, R0
	MOVBU	c+0(FP), R1

	SMC	$0

	RET

// func exit()
TEXT ·exit(SB),$0
	MOVD	$const_SYS_EXIT, R0

	SMC	$0

	RET

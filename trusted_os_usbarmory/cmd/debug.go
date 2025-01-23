// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"debug/elf"
	"encoding/binary"
	"fmt"
	"unsafe"

	"golang.org/x/term"

	layout "github.com/usbarmory/GoTEE-example/mem"
	"github.com/usbarmory/GoTEE-example/util"
)

func init() {
	Add(Cmd{
		Name: "allgptr",
		Help: "memory forensics of applet goroutines",
		Fn:   allgptrCmd,
	})
}

type m struct {
	g0      *g
	morebuf gobuf
}

type gobuf struct {
	sp   uint32
	pc   uint32
	g    *g
	ctxt uint32
	ret  uint32
	lr   uint32
	bp   uint32
}

type stack struct {
	lo uintptr
	hi uintptr
}

type g struct {
	stack            stack
	stackguard0      uintptr
	stackguard1      uintptr
	_panic           uintptr
	_defer           uintptr
	m                *m
	sched            gobuf
	syscallsp        uintptr
	syscallpc        uintptr
	stktopsp         uintptr
	param            uint
	atomicstatus     uint
	stackLock        uint32
	goid             uint64
	schedlink        uintptr
	waitsince        int64
	waitreason       uint8
	preempt          bool
	preemptStop      bool
	preemptShrink    bool
	asyncSafePoint   bool
	paniconfault     bool
	gcscandone       bool
	throwsplit       bool
	activeStackChans bool

	noCopy struct{}
	value  uint8

	raceignore     int8
	sysblocktraced bool
	tracking       bool
	trackingSeq    uint8
	trackingStamp  int64
	runnableTime   int64
	sysexitticks   int64
	traceseq       uint64
	tracelastp     uintptr
	lockedm        uintptr
	sig            uint32
	writebuf       []byte
	sigcode0       uintptr
	sigcode1       uintptr
	sigpc          uintptr
	gopc           uintptr
	ancestors      uintptr
	startpc        uintptr
}

func withinAppletMemory(ptr uint32) bool {
	return (ptr >= layout.AppletVirtualStart && ptr <= (layout.AppletVirtualStart+layout.AppletSize))
}

// allgptrCmd forensically profiles goroutines from Go runtime memory, this
// allows the inspection of state from a separate Go runtime (i.e. isolated
// from the one running this command) even after a warm reboot.
//
// The technique involves following the runtime.allgptr symbol to parse profile
// information from memory.
func allgptrCmd(term *term.Terminal, _ []string) (res string, err error) {
	var sym *elf.Symbol

	if sym, err = util.LookupSym("runtime.allgptr"); err != nil {
		return "", fmt.Errorf("could not find runtime.allgptr symbol, %v", err)
	}

	allgptr := (*uint32)(unsafe.Pointer(uintptr(sym.Value)))

	if !withinAppletMemory(*allgptr) {
		return "", fmt.Errorf("invalid allgptr (%x)", *allgptr)
	}

	if sym, err = util.LookupSym("runtime.allglen"); err != nil {
		return "", fmt.Errorf("could not find runtime.allglen symbol, %v", err)
	}

	allglen := (*uint32)(unsafe.Pointer(uintptr(sym.Value)))

	if sym, err = util.LookupSym("runtime.text"); err != nil {
		return "", fmt.Errorf("could not find runtime.text symbol, %v", err)
	}

	text := sym.Value

	if sym, err = util.LookupSym("runtime.etext"); err != nil {
		return "", fmt.Errorf("could not find runtime.etext symbol, %v", err)
	}

	etext := sym.Value

	for i := uint32(0); i < *allglen; i++ {
		gptr := (*uint32)(unsafe.Pointer(uintptr(*allgptr + i*4)))

		if !withinAppletMemory(*gptr) {
			fmt.Fprintf(term, "invalid gptr (%x)", *gptr)
			continue
		}

		g := (*g)(unsafe.Pointer(uintptr(*gptr)))

		fmt.Fprintf(term, "\ng[%d]: stack.lo:%x stack.hi:%x m:%x sched.sp:%x sched.pc:%x\n", i, g.stack.lo, g.stack.hi, g.m, g.sched.sp, g.sched.pc)

		if l, err := util.PCToLine(uint64(g.gopc)); err == nil {
			fmt.Fprintf(term, "\tgopc (%x): %s", g.gopc, l)
		}

		if g.m == nil {
			fmt.Fprintf(term, "\n")
		} else {
			fmt.Fprintf(term, " - goroutine was active, sweeping stack pointers\n")

			stack := mem(uint(g.stack.lo), int(g.stack.hi-g.stack.lo), nil)

			for i := 0; i < len(stack); i += 4 {
				try := uint64(binary.LittleEndian.Uint32(stack[i : i+4]))

				if try >= text && try <= etext {
					if l, err := util.PCToLine(try); err == nil {
						fmt.Fprintf(term, "\t%x\t%s\n", try, l)
					}
				}
			}
		}
	}

	return "", nil
}

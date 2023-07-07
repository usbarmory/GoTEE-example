// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
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
		Fn:   stackallptrCmd,
	})
}

type m struct {
	g0      *g
	morebuf gobuf
}

type gobuf struct {
	sp   uint32
	pc   uint32
	g    uint32
	ctxt uint32
	ret  uint32
	lr   uint32
	bp   uint32
}

type g struct {
	stacklo     uint32
	stackhi     uint32
	stackguard0 uint32
	stackguard1 uint32
	_panic      uint32
	_defer      uint32
	m           *m
	sched       gobuf
	syscallsp   uint32
	syscallpc   uint32
}

func withinAppletMemory(ptr uint32) bool {
	return (ptr >= layout.AppletStart && ptr <= (layout.AppletStart + layout.AppletSize))
}

func stackallptrCmd(term *term.Terminal, _ []string) (res string, err error) {
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
		gptr := (*uint32)(unsafe.Pointer(uintptr(*allgptr+i*4)))

		if !withinAppletMemory(*gptr) {
			fmt.Fprintf(term, "invalid gptr (%x)", *gptr)
			continue
		}

		g := (*g)(unsafe.Pointer(uintptr(*gptr)))

		fmt.Fprintf(term, "\ng[%d]: %x\n", i, g)

		if g.m == nil {
			if l, err := util.PCToLine(uint64(g.sched.pc)); err == nil {
				fmt.Fprintf(term, "\tg[%d].sched.pc (%x): %s\n", i, g.sched.pc, l)
			} else {
				fmt.Fprintf(term, "\tg[%d].sched: %x\n", i, g.sched)
			}
		} else {
			stack := mem(uint(g.stacklo), int(g.stackhi - g.stacklo), nil)

			for i := 0; i < len(stack); i += 4 {
				try := uint64(binary.LittleEndian.Uint32(stack[i:i+4]))

				if try >= text && try <= etext {
					if l, err := util.PCToLine(try); err == nil {
						fmt.Fprintf(term, "\tpotential LR (%x): %s\n", try, l)
					}
				}
			}
		}
	}

	return "", nil
}

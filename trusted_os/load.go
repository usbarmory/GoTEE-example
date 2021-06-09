// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	_ "embed"
	"log"
	"sync"
	_ "unsafe"

	"github.com/f-secure-foundry/GoTEE/monitor"
	"github.com/f-secure-foundry/GoTEE/syscall"

	"github.com/f-secure-foundry/tamago/arm"

	"github.com/f-secure-foundry/GoTEE-example/mem"

	"golang.org/x/term"
)

// This example simply embeds Trusted Applet and Main OS ELF binaries within
// the Trusted OS executable.
//
// The loading strategy is up to implementers, on the NXP i.MX6 the armory-boot
// bootloader primitives can be used to create a bootable Trusted OS with
// authenticated disk loading of applets and kernels, see:
//   https://pkg.go.dev/github.com/f-secure-foundry/armory-boot

//go:embed trusted_applet.elf
var taELF []byte

//go:embed nonsecure_os_go.elf
var osELF []byte

var secureOutput bytes.Buffer
var nonSecureOutput bytes.Buffer

const outputLimit = 1024

func sysWrite(buf *bytes.Buffer, c byte, color []byte, t *term.Terminal) {
	buf.WriteByte(c)

	if c == 0x0a || buf.Len() > outputLimit {
		t.Write(color)
		t.Write(buf.Bytes())
		t.Write(t.Escape.Reset)

		buf.Reset()
	}
}

func loadApplet() (ta *monitor.ExecCtx) {
	var err error

	if ta, err = monitor.Load(taELF, mem.AppletStart, mem.AppletSize, true); err != nil {
		log.Fatalf("PL1 could not load applet, %v", err)
	} else {
		log.Printf("PL1 loaded applet addr:%#x size:%d entry:%#x", ta.Memory.Start, len(taELF), ta.R15)
	}

	// register example RPC receiver
	ta.Server.Register(&RPC{})
	ta.Debug = true

	if ssh != nil {
		// When running on real hardware, override the default handler
		// to log applet stdout on SSH console.
		//
		// A buffer is used to flush only complete logs.

		ta.Handler = func(ctx *monitor.ExecCtx) error {
			if ctx.R0 == syscall.SYS_WRITE {
				sysWrite(&secureOutput, byte(ctx.R1), ssh.Term.Escape.Green, ssh.Term)
			}

			return monitor.SecureHandler(ctx)
		}
	}

	return
}

func loadNormalWorld(lock bool) (os *monitor.ExecCtx) {
	var err error

	if os, err = monitor.Load(osELF, mem.NonSecureStart, mem.NonSecureSize, false); err != nil {
		log.Fatalf("PL1 could not load applet, %v", err)
	} else {
		log.Printf("PL1 loaded kernel addr:%#x size:%d entry:%#x", os.Memory.Start, len(osELF), os.R15)
	}

	os.Debug = true

	if err = configureTrustZone(mem.NonSecureStart, mem.NonSecureSize, lock); err != nil {
		log.Fatalf("PL1 could not configure TrustZone, %v", err)
	}

	if ssh != nil {
		// When running on real hardware, override the default handler
		// to log NonSecure World stdout on Secure World SSH console.
		//
		// A buffer is used to flush only complete logs.

		os.Handler = func(ctx *monitor.ExecCtx) (err error) {
			if ctx.R0 == syscall.SYS_WRITE {
				sysWrite(&nonSecureOutput, byte(ctx.R1), ssh.Term.Escape.Red, ssh.Term)
			} else {
				err = monitor.NonSecureHandler(ctx)
			}

			return
		}
	}

	return
}

func run(ctx *monitor.ExecCtx, wg *sync.WaitGroup) {
	mode := arm.ModeName(int(ctx.SPSR) & 0x1f)
	ns := ctx.NonSecure()

	log.Printf("PL1 starting mode:%s ns:%v sp:%#.8x pc:%#.8x", mode, ns, ctx.R13, ctx.R15)

	err := ctx.Run()

	if wg != nil {
		wg.Done()
	}

	log.Printf("PL1 stopped mode:%s ns:%v sp:%#.8x lr:%#.8x pc:%#.8x err:%v", mode, ns, ctx.R13, ctx.R14, ctx.R15, err)
}

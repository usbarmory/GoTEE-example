// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	_ "embed"
	"fmt"
	"log"
	"sync"

	"github.com/f-secure-foundry/tamago/arm"

	"github.com/f-secure-foundry/GoTEE/monitor"
	"github.com/f-secure-foundry/GoTEE/syscall"

	"github.com/f-secure-foundry/GoTEE-example/mem"
)

// This example embeds the Trusted Applet and Main OS ELF binaries within the
// Trusted OS executable, using Go embed package.
//
// The loading strategy is up to implementers, on the NXP i.MX6 the armory-boot
// bootloader primitives can be used to create a bootable Trusted OS with
// authenticated disk loading of applets and kernels, see:
//   https://pkg.go.dev/github.com/f-secure-foundry/armory-boot

//go:embed assets/trusted_applet.elf
var taELF []byte

//go:embed assets/nonsecure_os_go.elf
var osELF []byte

func loadApplet() (ta *monitor.ExecCtx, err error) {
	if ta, err = monitor.Load(taELF, mem.AppletStart, mem.AppletSize, true); err != nil {
		return nil, fmt.Errorf("PL1 could not load applet, %v", err)
	} else {
		log.Printf("PL1 loaded applet addr:%#x size:%d entry:%#x", ta.Memory.Start, len(taELF), ta.R15)
	}

	// register example RPC receiver
	ta.Server.Register(&RPC{})
	ta.Debug = true

	// set stack pointer to the end of applet memory
	ta.R13 = mem.AppletStart + mem.AppletSize

	// The GoTEE default handler is overridden to avoid interleaved logs,
	// as the supervisor and applet contexts are logging simultaneously.
	//
	// When running on real hardware logs are cloned on the SSH terminal.
	ta.Handler = func(ctx *monitor.ExecCtx) (err error) {
		if ctx.R0 == syscall.SYS_WRITE {
			if ssh != nil {
				bufferedTermLog(&secureOutput, byte(ctx.R1), ssh.Term.Escape.Green, ssh.Term)
			} else {
				bufferedStdoutLog(&secureOutput, byte(ctx.R1))
			}
		} else {
			err = monitor.SecureHandler(ctx)
		}

		return
	}

	return
}

func loadNormalWorld(lock bool) (os *monitor.ExecCtx, err error) {
	if os, err = monitor.Load(osELF, mem.NonSecureStart, mem.NonSecureSize, false); err != nil {
		return nil, fmt.Errorf("PL1 could not load applet, %v", err)
	} else {
		log.Printf("PL1 loaded kernel addr:%#x size:%d entry:%#x", os.Memory.Start, len(osELF), os.R15)
	}

	os.Debug = true

	if err = configureTrustZone(mem.NonSecureStart, mem.NonSecureSize, lock); err != nil {
		return nil, fmt.Errorf("PL1 could not configure TrustZone, %v", err)
	}

	// The GoTEE default handler is overridden to avoid interleaved logs,
	// as the supervisor and nonsecure contexts are logging simultaneously.
	//
	// When running on real hardware logs are cloned on the SSH terminal.
	os.Handler = func(ctx *monitor.ExecCtx) (err error) {
		if ctx.R0 == syscall.SYS_WRITE {
			if ssh != nil {
				bufferedTermLog(&nonSecureOutput, byte(ctx.R1), ssh.Term.Escape.Red, ssh.Term)
			} else {
				bufferedStdoutLog(&nonSecureOutput, byte(ctx.R1))
			}
		} else {
			err = monitor.NonSecureHandler(ctx)
		}

		return
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

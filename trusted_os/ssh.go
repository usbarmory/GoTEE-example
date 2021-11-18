// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"runtime/debug"
	"runtime/pprof"
	"strconv"

	"golang.org/x/term"

	"github.com/f-secure-foundry/tamago/arm"
	"github.com/f-secure-foundry/tamago/bits"
	"github.com/f-secure-foundry/tamago/board/f-secure/usbarmory/mark-two"
	"github.com/f-secure-foundry/tamago/dma"
	"github.com/f-secure-foundry/tamago/soc/imx6"
	"github.com/f-secure-foundry/tamago/soc/imx6/csu"
)

const MD_LIMIT = 102400

const help = `
  help                                   # this help
  reboot                                 # reset the SoC/board
  stack                                  # stack trace of current goroutine
  stackall                               # stack trace of all goroutines
  md  <hex offset> <size>                # memory display (use with caution)
  mw  <hex offset> <hex value>           # memory write   (use with caution)

  gotee                                  # TrustZone test w/ TamaGo unikernels
  linux <uSD|eMMC>                       # Boot Non-secure USB armory Debian base image

  dbg                                    # show ARM debug permissions
  csl                                    # show config security levels (CSL)
  csl <periph> <slave> <hex csl>         #  set config security level  (CSL)
  sa                                     # show security access (SA)
  sa  <id> <secure|nonsecure>            #  set security access (SA)

`

var memoryCommandPattern = regexp.MustCompile(`(md|mw) ([[:xdigit:]]+) (\d+|[[:xdigit:]]+)`)
var cslCommandPattern = regexp.MustCompile(`csl (\d+) (\d+) ([[:xdigit:]]+)`)
var saCommandPattern = regexp.MustCompile(`sa (\d+) (secure|nonsecure)`)
var linuxCommandPattern = regexp.MustCompile(`linux (uSD|eMMC)`)

func memAccess(start uint32, size int, w []byte) (b []byte) {
	// temporarily map page zero if required
	if z := uint32(1 << 20); start < z {
		csu.SetAccess(0, true, false)

		imx6.ARM.ConfigureMMU(0, z, (arm.TTE_AP_001<<10)|arm.TTE_SECTION)
		defer imx6.ARM.ConfigureMMU(0, z, 0)
	}

	mem := &dma.Region{
		Start: uint32(start),
		Size:  size,
	}
	mem.Init()

	start, buf := mem.Reserve(size, 0)
	defer mem.Release(start)

	if len(w) > 0 {
		copy(buf, w)
	} else {
		b = make([]byte, size)
		copy(b, buf)
	}

	return
}

func memoryCommand(arg []string) (res string) {
	addr, err := strconv.ParseUint(arg[1], 16, 32)

	if err != nil {
		return fmt.Sprintf("invalid address: %v", err)
	}

	switch arg[0] {
	case "md":
		size, err := strconv.ParseUint(arg[2], 10, 32)

		if err != nil {
			return fmt.Sprintf("invalid size: %v", err)
		}

		if (addr%4) != 0 || (size%4) != 0 {
			return "please only perform 32-bit aligned accesses"
		}

		if size > MD_LIMIT {
			return fmt.Sprintf("please only use a size argument <= %d", MD_LIMIT)
		}

		return hex.Dump(memAccess(uint32(addr), int(size), nil))
	case "mw":
		val, err := strconv.ParseUint(arg[2], 16, 32)

		if err != nil {
			return fmt.Sprintf("invalid data: %v", err)
		}

		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(val))

		memAccess(uint32(addr), 4, buf)
	}

	return
}

func cslCommand(arg []string) (res string) {
	if arg == nil {
		var buf bytes.Buffer

		for i := csu.CSL_MIN; i < csu.CSL_MAX; i++ {
			csl, _, _ := csu.GetSecurityLevel(i, 0)
			buf.WriteString(fmt.Sprintf("CSL%.2d 0:%#.2x", i, csl))

			csl, _, _ = csu.GetSecurityLevel(i, 1)
			buf.WriteString(fmt.Sprintf(" 1:%#.2x\n", csl))
		}

		return buf.String()
	}

	periph, err := strconv.ParseUint(arg[0], 10, 8)

	if err != nil {
		return fmt.Sprintf("invalid peripheral index: %v", err)
	}

	slave, err := strconv.ParseUint(arg[1], 10, 8)

	if err != nil {
		return fmt.Sprintf("invalid slave index: %v", err)
	}

	csl, err := strconv.ParseUint(arg[2], 16, 8)

	if err != nil {
		return fmt.Sprintf("invalid csl: %v", err)
	}

	err = csu.SetSecurityLevel(int(periph), int(slave), uint8(csl), false)

	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	return
}

func saCommand(arg []string) (res string) {
	if arg == nil {
		var buf bytes.Buffer

		for i := csu.SA_MIN; i < csu.SA_MAX; i++ {
			if sa, _, _ := csu.GetAccess(i); sa {
				buf.WriteString(fmt.Sprintf("SA%.2d: secure\n", i))
			} else {
				buf.WriteString(fmt.Sprintf("SA%.2d: nonsecure\n", i))
			}
		}

		return buf.String()
	}

	id, err := strconv.ParseUint(arg[0], 10, 8)

	if err != nil {
		return fmt.Sprintf("invalid peripheral index: %v", err)
	}

	if arg[1] == "secure" {
		err = csu.SetAccess(int(id), true, false)
	} else {
		err = csu.SetAccess(int(id), false, false)
	}

	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	return
}

func linuxCommand(arg []string) (res string) {
	if err := linux(arg[0]); err != nil {
		return fmt.Sprintf("%v", err)
	}

	return
}

func dbg() string {
	var buf bytes.Buffer

	dbgAuthStatus := imx6.ARM.DebugStatus()

	buf.WriteString("| type                    | implemented | enabled |\n")
	buf.WriteString("|-------------------------|-------------|---------|\n")

	buf.WriteString(fmt.Sprintf("| Secure non-invasive     |           %d |       %d |\n",
		bits.Get(&dbgAuthStatus, 7, 1),
		bits.Get(&dbgAuthStatus, 6, 1),
	))

	buf.WriteString(fmt.Sprintf("| Secure invasive         |           %d |       %d |\n",
		bits.Get(&dbgAuthStatus, 5, 1),
		bits.Get(&dbgAuthStatus, 4, 1),
	))

	buf.WriteString(fmt.Sprintf("| Non-secure non-invasive |           %d |       %d |\n",
		bits.Get(&dbgAuthStatus, 3, 1),
		bits.Get(&dbgAuthStatus, 2, 1),
	))

	buf.WriteString(fmt.Sprintf("| Non-secure invasive     |           %d |       %d |\n",
		bits.Get(&dbgAuthStatus, 1, 1),
		bits.Get(&dbgAuthStatus, 0, 1),
	))

	return buf.String()
}

func cmd(term *term.Terminal, cmd string) (err error) {
	var res string

	switch cmd {
	case "exit", "quit":
		res = "logout"
		err = io.EOF
	case "help":
		res = string(term.Escape.Cyan) + help + string(term.Escape.Reset)
	case "reboot":
		usbarmory.Reset()
	case "stack":
		res = string(debug.Stack())
	case "stackall":
		buf := new(bytes.Buffer)
		pprof.Lookup("goroutine").WriteTo(buf, 1)
		res = buf.String()
	case "gotee":
		err = gotee()
	case "dbg":
		res = dbg()
	case "csl":
		res = cslCommand(nil)
	case "sa":
		res = saCommand(nil)
	default:
		if m := memoryCommandPattern.FindStringSubmatch(cmd); len(m) == 4 {
			res = memoryCommand(m[1:])
		} else if m := cslCommandPattern.FindStringSubmatch(cmd); len(m) == 4 {
			res = cslCommand(m[1:])
		} else if m := saCommandPattern.FindStringSubmatch(cmd); len(m) == 3 {
			res = saCommand(m[1:])
		} else if m := linuxCommandPattern.FindStringSubmatch(cmd); len(m) == 2 {
			res = linuxCommand(m[1:])
		} else {
			res = "unknown command, type `help`"
		}
	}

	fmt.Fprintln(term, res)

	return
}

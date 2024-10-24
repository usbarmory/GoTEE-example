// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/nxp/csu"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

func init() {
	Add(Cmd{
		Name: "csl",
		Help: "show config security levels (CSL)",
		Fn:   cslCmd,
	})

	Add(Cmd{
		Name:    "csl ",
		Args:    3,
		Pattern: regexp.MustCompile(`csl (\d+) (\d+) ([[:xdigit:]]+)`),
		Syntax:  "<periph> <slave> <hex csl>",
		Help:    "set config security level (CSL)",
		Fn:      cslCmd,
	})

	Add(Cmd{
		Name: "sa",
		Help: "show security access (SA)",
		Fn:   saCmd,
	})

	Add(Cmd{
		Name:    "sa ",
		Args:    2,
		Pattern: regexp.MustCompile(`sa (\d+) (secure|nonsecure)`),
		Syntax:  "<id> <secure|nonsecure>",
		Help:    "set security access (SA)",
		Fn:      saCmd,
	})
}

func cslCmd(term *term.Terminal, arg []string) (res string, err error) {
	if len(arg) == 0 {
		var buf bytes.Buffer

		for i := csu.CSL_MIN; i <= csu.CSL_MAX; i++ {
			csl, _, _ := imx6ul.CSU.GetSecurityLevel(i, 0)
			fmt.Fprintf(&buf, "CSL%.2d 0:%#.2x", i, csl)

			csl, _, _ = imx6ul.CSU.GetSecurityLevel(i, 1)
			fmt.Fprintf(&buf, " 1:%#.2x\n", csl)
		}

		return buf.String(), nil
	}

	if !imx6ul.Native {
		return "", errors.New("unsupported under emulation")
	}

	periph, err := strconv.ParseUint(arg[0], 10, 8)

	if err != nil {
		return "", fmt.Errorf("invalid peripheral index: %v", err)
	}

	slave, err := strconv.ParseUint(arg[1], 10, 8)

	if err != nil {
		return "", fmt.Errorf("invalid slave index: %v", err)
	}

	csl, err := strconv.ParseUint(arg[2], 16, 8)

	if err != nil {
		return "", fmt.Errorf("invalid csl: %v", err)
	}

	if err = imx6ul.CSU.SetSecurityLevel(int(periph), int(slave), uint8(csl), false); err != nil {
		return
	}

	return
}

func saCmd(term *term.Terminal, arg []string) (res string, err error) {
	if len(arg) == 0 {
		var buf bytes.Buffer

		for i := csu.SA_MIN; i <= csu.SA_MAX; i++ {
			if sa, _, _ := imx6ul.CSU.GetAccess(i); sa {
				fmt.Fprintf(&buf, "SA%.2d: secure\n", i)
			} else {
				fmt.Fprintf(&buf, "SA%.2d: nonsecure\n", i)
			}
		}

		return buf.String(), nil
	}

	if !imx6ul.Native {
		return "", errors.New("unsupported under emulation")
	}

	id, err := strconv.ParseUint(arg[0], 10, 8)

	if err != nil {
		return "", fmt.Errorf("invalid peripheral index: %v", err)
	}

	if arg[1] == "secure" {
		err = imx6ul.CSU.SetAccess(int(id), true, false)
	} else {
		err = imx6ul.CSU.SetAccess(int(id), false, false)
	}

	return
}

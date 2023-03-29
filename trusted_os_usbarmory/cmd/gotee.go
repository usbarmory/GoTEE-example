// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"regexp"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/nxp/imx6ul"

	"github.com/usbarmory/GoTEE-example/trusted_os_usbarmory/internal"
)

func init() {
	Add(Cmd{
		Name: "gotee",
		Help: "TrustZone example w/ TamaGo unikernels",
		Fn:   goteeCmd,
	})

	Add(Cmd{
		Name:    "linux",
		Args:    1,
		Pattern: regexp.MustCompile(`^linux (uSD|eMMC)$`),
		Syntax:  "<uSD|eMMC>",
		Help:    "boot NonSecure USB armory Debian base image",
		Fn:      linuxCmd,
	})
}

func goteeCmd(term *term.Terminal, arg []string) (res string, err error) {
	return "", gotee.GoTEE()
}

func linuxCmd(term *term.Terminal, arg []string) (res string, err error) {
	if !imx6ul.Native {
		return "", errors.New("unsupported under emulation")
	}

	return "", gotee.Linux(arg[0])
}

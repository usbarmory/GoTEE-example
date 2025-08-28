// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build sifive_u
// +build sifive_u

package cmd

import (
	"fmt"
	"regexp"
	"strconv"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/sifive/fu540"
)

func init() {
	Add(Cmd{
		Name:    "pmp ",
		Args:    1,
		Pattern: regexp.MustCompile(`^pmp (\d+)$`),
		Syntax:  "<index>",
		Help:    "read PMP CSR",
		Fn:      pmpRead,
	})

	Add(Cmd{
		Name:    "pmp",
		Args:    7,
		Pattern: regexp.MustCompile(`^pmp (\d+) ([[:xdigit:]]+) (\S+) (\S+) (\S+) (\d+) (\S+)$`),
		Syntax:  "<index> <hex addr> <a> <r> <w> <x> <l>",
		Help:    "write PMP CSR",
		Fn:      pmpWrite,
	})
}

func pmpRead(_ *term.Terminal, arg []string) (res string, err error) {
	i, err := strconv.ParseUint(arg[0], 10, 8)

	if err != nil {
		return "", fmt.Errorf("invalid index, %v", err)
	}

	addr, r, w, x, a, l, err := fu540.RV64.ReadPMP(int(i))

	if err != nil {
		return
	}

	return fmt.Sprintf("PMP:%.2d addr:%.16x A:%d R:%v W:%v X:%v l:%v", i, addr, a, r, w, x, l), nil
}

func pmpWrite(_ *term.Terminal, arg []string) (res string, err error) {
	i, err := strconv.ParseUint(arg[0], 10, 8)

	if err != nil {
		return "", fmt.Errorf("invalid index, %v", err)
	}

	addr, err := strconv.ParseUint(arg[1], 16, 64)

	if err != nil {
		return "", fmt.Errorf("invalid address, %v", err)
	}

	a, err := strconv.ParseUint(arg[2], 10, 2)

	if err != nil {
		return "", fmt.Errorf("invalid A value, %v", err)
	}

	r, err := strconv.ParseBool(arg[3])

	if err != nil {
		return "", fmt.Errorf("invalid R boolean, %v", err)
	}

	w, err := strconv.ParseBool(arg[4])

	if err != nil {
		return "", fmt.Errorf("invalid W boolean, %v", err)
	}

	x, err := strconv.ParseBool(arg[5])

	if err != nil {
		return "", fmt.Errorf("invalid X boolean, %v", err)
	}

	l, err := strconv.ParseBool(arg[6])

	if err != nil {
		return "", fmt.Errorf("invalid l boolean, %v", err)
	}

	err = fu540.RV64.WritePMP(int(i), addr, r, w, x, int(a), l)

	return
}

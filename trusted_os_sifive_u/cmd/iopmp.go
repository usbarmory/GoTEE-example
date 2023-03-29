// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build sifive_u
// +build sifive_u

package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/sifive/physicalfilter"
)

var Base uint32

func init() {
	Add(Cmd{
		Name:    "iopmp",
		Args:    1,
		Pattern: regexp.MustCompile(`^iopmp (\d+)$`),
		Syntax:  "<index>",
		Help:    "read Device PMP",
		Fn:      iopmpRead,
	})
}

func iopmpRead(_ *term.Terminal, arg []string) (res string, err error) {
	if Base == 0 {
		return "", errors.New("unavailable")
	}

	i, err := strconv.ParseUint(arg[0], 10, 8)

	if err != nil {
		return "", fmt.Errorf("invalid size, %v", err)
	}

	pf := &physicalfilter.PhysicalFilter{
		Base: Base,
	}

	addr, r, w, a, lock, err := pf.ReadPMP(int(i))

	if err != nil {
		return
	}

	return fmt.Sprintf("DevicePMP:%.2d addr:%#.16x A:%v R:%v W:%v lock:%v", i, addr, a, r, w, lock), nil
}

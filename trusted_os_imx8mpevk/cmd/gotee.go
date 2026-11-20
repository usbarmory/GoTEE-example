// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"golang.org/x/term"

	"github.com/usbarmory/GoTEE-example/trusted_os_imx8mpevk/internal"
)

func init() {
	Add(Cmd{
		Name: "gotee",
		Help: "EL0 example w/ TamaGo unikernels",
		Fn:   goteeCmd,
	})
}

func goteeCmd(term *term.Terminal, arg []string) (res string, err error) {
	return "", gotee.GoTEE()
}

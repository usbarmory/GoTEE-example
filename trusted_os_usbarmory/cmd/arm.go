// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"errors"
	"fmt"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

func init() {
	Add(Cmd{
		Name: "dbg",
		Help: "show ARM debug permissions",
		Fn:   dbgCmd,
	})
}

func dbgCmd(term *term.Terminal, arg []string) (res string, err error) {
	var buf bytes.Buffer

	if !imx6ul.Native {
		return "", errors.New("unsupported under emulation")
	}

	dbgAuthStatus := imx6ul.ARM.DebugStatus()

	buf.WriteString("| type                    | implemented | enabled |\n")
	buf.WriteString("|-------------------------|-------------|---------|\n")

	fmt.Fprintf(&buf, "| Secure non-invasive     |           %d |       %d |\n",
		bits.GetN(&dbgAuthStatus, 7, 1),
		bits.GetN(&dbgAuthStatus, 6, 1),
	)

	fmt.Fprintf(&buf, "| Secure invasive         |           %d |       %d |\n",
		bits.GetN(&dbgAuthStatus, 5, 1),
		bits.GetN(&dbgAuthStatus, 4, 1),
	)

	fmt.Fprintf(&buf, "| Non-secure non-invasive |           %d |       %d |\n",
		bits.GetN(&dbgAuthStatus, 3, 1),
		bits.GetN(&dbgAuthStatus, 2, 1),
	)

	fmt.Fprintf(&buf, "| Non-secure invasive     |           %d |       %d |\n",
		bits.GetN(&dbgAuthStatus, 1, 1),
		bits.GetN(&dbgAuthStatus, 0, 1),
	)

	return buf.String(), nil
}

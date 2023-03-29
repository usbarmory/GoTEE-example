// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package util

import (
	"bytes"
	"os"

	"golang.org/x/term"
)

var secureOutput bytes.Buffer
var nonSecureOutput bytes.Buffer

const outputLimit = 1024
const flushChr = 0x0a // \n

func BufferedStdoutLog(c byte, secure bool) {
	var buf *bytes.Buffer

	if secure {
		buf = &secureOutput
	} else {
		buf = &nonSecureOutput
	}

	buf.WriteByte(c)

	if c == flushChr || buf.Len() > outputLimit {
		os.Stdout.Write(buf.Bytes())
		buf.Reset()
	}
}

func BufferedTermLog(c byte, secure bool, t *term.Terminal) {
	var buf *bytes.Buffer
	var color []byte

	if secure {
		buf = &secureOutput
		color = t.Escape.Green
	} else {
		buf = &nonSecureOutput
		color = t.Escape.Red
	}

	buf.WriteByte(c)

	if c == flushChr || buf.Len() > outputLimit {
		t.Write(color)
		t.Write(buf.Bytes())
		t.Write(t.Escape.Reset)

		buf.Reset()
	}
}

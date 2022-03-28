// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"os"

	"golang.org/x/term"
)

var secureOutput bytes.Buffer
var nonSecureOutput bytes.Buffer

const outputLimit = 1024
const flushChr = 0x0a // \n

func bufferedStdoutLog(c byte, ns bool) {
	var buf *bytes.Buffer

	if ns {
		buf = &nonSecureOutput
	} else {
		buf = &secureOutput
	}

	buf.WriteByte(c)

	if c == flushChr || buf.Len() > outputLimit {
		os.Stdout.Write(buf.Bytes())
		buf.Reset()
	}
}

func bufferedTermLog(c byte, ns bool, t *term.Terminal) {
	var buf *bytes.Buffer
	var color []byte

	if ns {
		buf = &nonSecureOutput
		color = t.Escape.Red
	} else {
		buf = &secureOutput
		color = t.Escape.Green
	}

	buf.WriteByte(c)

	if c == flushChr || buf.Len() > outputLimit {
		os.Stdout.Write(buf.Bytes())

		t.Write(color)
		t.Write(buf.Bytes())
		t.Write(t.Escape.Reset)

		buf.Reset()
	}
}

// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
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

func bufferedStdoutLog(buf *bytes.Buffer, c byte) {
	buf.WriteByte(c)

	if c == flushChr || buf.Len() > outputLimit {
		os.Stdout.Write(buf.Bytes())
		buf.Reset()
	}
}

func bufferedTermLog(buf *bytes.Buffer, c byte, color []byte, t *term.Terminal) {
	buf.WriteByte(c)

	if c == flushChr || buf.Len() > outputLimit {
		os.Stdout.Write(buf.Bytes())

		t.Write(color)
		t.Write(buf.Bytes())
		t.Write(t.Escape.Reset)

		buf.Reset()
	}
}

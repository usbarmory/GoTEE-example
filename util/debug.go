// Copyright 2022 The Armored Witness OS authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"bytes"
	"debug/elf"
	"debug/gosym"
	"errors"
	"fmt"
)

func LookupSym(buf []byte, name string) (*elf.Symbol, error) {
	exe, err := elf.NewFile(bytes.NewReader(buf))

	if err != nil {
		return nil, err
	}

	syms, err := exe.Symbols()

	if err != nil {
		return nil, err
	}

	for _, sym := range syms {
		if sym.Name == name {
			return &sym, nil
		}
	}

	return nil, errors.New("symbol not found")
}

func goSymTable(buf []byte) (symTable *gosym.Table, err error) {
	exe, err := elf.NewFile(bytes.NewReader(buf))

	if err != nil {
		return
	}

	addr := exe.Section(".text").Addr

	lineTableData, err := exe.Section(".gopclntab").Data()

	if err != nil {
		return
	}

	lineTable := gosym.NewLineTable(lineTableData, addr)

	if err != nil {
		return
	}

	symTableData, err := exe.Section(".gosymtab").Data()

	if err != nil {
		return
	}

	return gosym.NewTable(symTableData, lineTable)
}

func PCToLine(buf []byte, pc uint64) (s string, err error) {
	symTable, err := goSymTable(buf)

	if err != nil {
		return
	}

	file, line, _ := symTable.PCToLine(pc)

	return fmt.Sprintf("%s:%d", file, line), nil
}

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

var (
	target        []byte
	symCache      []elf.Symbol
	symTableCache *gosym.Table
)

func SetDebugTarget(buf []byte) {
	target = buf
}

func LookupSym(name string) (*elf.Symbol, error) {
	f, err := elf.NewFile(bytes.NewReader(target))

	if err != nil {
		return nil, err
	}

	if symCache == nil {
		syms, err := f.Symbols()

		if err != nil {
			return nil, err
		}

		symCache = syms
	}

	for _, sym := range symCache {
		if sym.Name == name {
			return &sym, nil
		}
	}

	return nil, errors.New("symbol not found")
}

func goSymTable() (symTable *gosym.Table, err error) {
	if symTableCache != nil {
		return symTableCache, nil
	}

	f, err := elf.NewFile(bytes.NewReader(target))

	if err != nil {
		return
	}

	addr := f.Section(".text").Addr

	lineTableData, err := f.Section(".gopclntab").Data()

	if err != nil {
		return
	}

	lineTable := gosym.NewLineTable(lineTableData, addr)

	if err != nil {
		return
	}

	symTableData, err := f.Section(".gosymtab").Data()

	if err != nil {
		return
	}

	symTableCache, err = gosym.NewTable(symTableData, lineTable)

	return symTableCache, err
}

func PCToLine(pc uint64) (s string, err error) {
	symTable, err := goSymTable()

	if err != nil {
		return
	}

	file, line, _ := symTable.PCToLine(pc)

	return fmt.Sprintf("%s:%d", file, line), nil
}

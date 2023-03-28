// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.
/*
 * Parts of this file were generated with makeClass --run. Edit only those parts of
 * the code inside of 'EXISTING_CODE' tags.
 */

package types

// EXISTING_CODE
import (
	"io"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
)

// EXISTING_CODE

type RawMonitorClean struct {
	Address  string `json:"address"`
	Dups     string `json:"dups"`
	SizeNow  string `json:"sizeNow"`
	SizeThen string `json:"sizeThen"`
	// EXISTING_CODE
	// EXISTING_CODE
}

type SimpleMonitorClean struct {
	Address  base.Address     `json:"address"`
	Dups     int64            `json:"dups"`
	SizeNow  int64            `json:"sizeNow"`
	SizeThen int64            `json:"sizeThen"`
	raw      *RawMonitorClean `json:"-"`
	// EXISTING_CODE
	// EXISTING_CODE
}

func (s *SimpleMonitorClean) Raw() *RawMonitorClean {
	return s.raw
}

func (s *SimpleMonitorClean) SetRaw(raw *RawMonitorClean) {
	s.raw = raw
}

func (s *SimpleMonitorClean) Model(showHidden bool, format string, extraOptions map[string]any) Model {
	var model = map[string]interface{}{}
	var order = []string{}

	// EXISTING_CODE
	model = map[string]any{
		"address":  s.Address,
		"sizeNow":  s.SizeNow,
		"sizeThen": s.SizeThen,
		"dups":     s.Dups,
	}
	order = []string{
		"address",
		"sizeNow",
		"sizeThen",
		"dups",
	}
	// EXISTING_CODE

	return Model{
		Data:  model,
		Order: order,
	}
}

func (s *SimpleMonitorClean) WriteTo(w io.Writer) (n int64, err error) {
	// EXISTING_CODE
	// EXISTING_CODE
	return 0, nil
}

func (s *SimpleMonitorClean) ReadFrom(r io.Reader) (n int64, err error) {
	// EXISTING_CODE
	// EXISTING_CODE
	return 0, nil
}

// EXISTING_CODE
// EXISTING_CODE

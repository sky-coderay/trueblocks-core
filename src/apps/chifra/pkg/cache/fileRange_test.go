// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.

package cache

import (
	"fmt"
	"testing"
)

type TestType struct {
	fileName string
	expected FileRange
	errored  bool
}

func TestFileRange(t *testing.T) {
	tests := []TestType{
		{
			fileName: "0.bin",
			expected: FileRange{0, 0},
		},
		{
			fileName: "illformed-x",
			expected: FileRange{0, 0},
			errored:  true,
		},
		{
			fileName: "9991-9909.bin",
			expected: FileRange{9991, 9909},
		},
		{
			fileName: "000100001-000100002.bin",
			expected: FileRange{100001, 100002},
		},
		{
			fileName: "891919121.bin",
			expected: FileRange{0, 891919121},
		},
		{
			fileName: "/unchained/mainnet/blooms/013433393-013436307.bloom",
			expected: FileRange{13433393, 13436307},
		},
	}

	check := func(t *testing.T, tt TestType, s FileRange, e error) {
		if tt.errored && e == nil {
			t.Error("Expected an error but got none", s)
		} else if !tt.errored && e != nil {
			t.Error("Unexpected error", e)
		}
	}

	for _, tt := range tests {
		s, err := RangeFromFilename(tt.fileName)
		check(t, tt, s, err)
		fmt.Println(s)
	}
}

func TestFilenameFromRange(t *testing.T) {
	type args struct {
		fileRange FileRange
		extension string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Pads numbers",
			args: args{
				fileRange: FileRange{
					First: 0,
					Last:  1,
				},
				extension: "bloom",
			},
			want: "000000000-000000001.bloom",
		},
		{
			name: "Random range 1",
			args: args{
				fileRange: FileRange{
					First: 14040187,
					Last:  14043115,
				},
				extension: "bloom",
			},
			want: "014040187-014043115.bloom",
		},
		{
			name: "Random range 2",
			args: args{
				fileRange: FileRange{
					First: 1371353,
					Last:  1504328,
				},
				extension: "bloom",
			},
			want: "001371353-001504328.bloom",
		},
		{
			name: "Works without extension",
			args: args{
				fileRange: FileRange{
					First: 1371353,
					Last:  1504328,
				},
				extension: "",
			},
			want: "001371353-001504328",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilenameFromRange(tt.args.fileRange, tt.args.extension); got != tt.want {
				t.Errorf("FilenameFromRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

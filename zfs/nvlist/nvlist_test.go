// Copyright 2018 Ayan George.
// All rights reserved.  Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvlist_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/ayang64/ztool/zfs/nvlist"
)

func BenchmarkRead(b *testing.B) {
	benches := []struct {
		Name   string
		Path   string
		Offset int64
		Size   int
	}{
		{Name: "", Path: "../../../test-data/uncompressed-simple.image", Offset: 0x4000, Size: 0x1c00},
	}

	for _, bench := range benches {
		b.Run(bench.Name, func(b *testing.B) {
			fh, err := os.Open(bench.Path)

			if err != nil {
				b.Fatal(err)
			}
			defer fh.Close()
			fh.Seek(bench.Offset, 0)
			nvl := make([]byte, bench.Size)

			fh.Read(nvl)
			br := bytes.NewReader(nvl)

			for i := 0; i < b.N; i++ {
				nvlist.Read(br)
				br.Reset(nvl)
			}
		})
	}
}

func TestLooper(t *testing.T) {
	tests := []struct {
		Name   string
		Path   string
		Offset int64
		Size   int64
		Opts   []func() func(*nvlist.Scanner) error
	}{
		{
			Name:   "OrigData",
			Path:   "../../../../zbackup-editable",
			Offset: 0x4000,
			Size:   0x1c00,
			Opts:   []func() func(*nvlist.Scanner) error{},
		},
		{
			Name:   "",
			Path:   "../../../test-data/uncompressed-simple.image",
			Offset: 0x4000,
			Size:   0x1c00,
			Opts:   []func() func(*nvlist.Scanner) error{},
		},
		{
			Name:   "",
			Path:   "../../../test-data/zpool.cache",
			Offset: 0x0,
			Size:   0x184c,
			Opts:   []func() func(*nvlist.Scanner) error{nvlist.WithoutHeader},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fh, err := os.Open(test.Path)
			if err != nil {
				t.Fatalf("could not open %q; %v", test.Path, err)
			}
			defer fh.Close()

			fh.Seek(test.Offset, 0)

			m, err := nvlist.Read(io.LimitReader(fh, test.Size))

			if err != nil {
				t.Fatal(err)
			}

			t.Run("ListFind", func(t *testing.T) {
				asize, found := m.Find("ashift")
				t.Logf("asize: %v, %v", asize, found)
			})

			o, err := json.MarshalIndent(m, "", "  ")

			if err != nil {
				t.Fatal(err)
			}

			t.Logf("\n%s", string(o))
		})
	}
}

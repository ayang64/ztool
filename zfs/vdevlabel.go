// Copyright 2018 Ayan George.
// All rights reserved.  Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zfs

type VdevLabel struct {
	BlankSpace      [8 << 10]byte   // 8k blank to accommodate os data
	BootBlockHeader [8 << 10]byte   // 8k reserved blank space
	NVP             [112 << 10]byte // XDR encoded  name value pairs
	UberBlockBuf    [128 << 10]byte // uber block array
}

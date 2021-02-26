// Copyright 2018 Ayan George.
// All rights reserved.  Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvlist

// ListMeta encodes the version and flags (if any) of an nvlist.
type ListMeta struct {
	Version int32
	Flags   uint32
}

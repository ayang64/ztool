// Copyright 2018 Ayan George.
// All rights reserved.  Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvlist

import "io"

// Read recursively parses the nvlist stored in the supplied
// io.Reader.  It is up to the caller to ensure that the
// reader is in position to start reading the nvlist.
func Read(r io.Reader, opts ...func(*Scanner) error) (List, error) {
	rc := make(List)
	scn := NewScanner(r, opts...)

	for scn.Next() {
		rc[scn.Name()] = scn.Value()
	}

	if err := scn.Error(); err != nil {
		return nil, err
	}
	return rc, nil
}

// Copyright 2018 Ayan George.
// All rights reserved.  Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvlist

type List map[string]interface{}

func (l List) Find(target string) (interface{}, bool) {
	// search submaps for value.
	for k := range l {
		sl, t := l[k].(List)
		if t == false {
			continue
		}
		if v, found := sl.Find(target); found {
			return v, found
		}
	}

	// search current level for value
	for k := range l {
		if k != target {
			continue
		}
		return l[k], true
	}

	return nil, false
}

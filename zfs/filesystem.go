package zfs

import "io"

type Filesystem struct {
	rs     io.ReadSeeker
	nvlist map[string]interface{}
	ub     *UberBlock
}

func WithNVList(nvl map[string]interface{}) func(*Filesystem) error {
	return func(fs *Filesystem) error {
		fs.nvlist = nvl
		return nil
	}
}

func WithReadSeeker(rs io.ReadSeeker) func(*Filesystem) error {
	return func(fs *Filesystem) error {
		fs.rs = rs
		return nil
	}
}

func New(opts ...func(*Filesystem) error) (*Filesystem, error) {
	rc := Filesystem{}
	for _, opt := range opts {
		if err := opt(&rc); err != nil {
			return nil, err
		}
	}
	return &rc, nil
}

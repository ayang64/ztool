package nvlist

// ListMeta encodes the version and flags (if any) of an nvlist.
type ListMeta struct {
	Version int32
	Flags   uint32
}

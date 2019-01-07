package nvlist

import "io"

func ReadFull(r io.Reader) (map[string]interface{}, error) {
	rc := make(map[string]interface{})
	scn := NewScanner(r)
	if err := scn.Error(); err != nil {
		return nil, err
	}

	for scn.Next() {
		rc[scn.Name()] = scn.Value()
	}

	if err := scn.Error(); err != nil {
		return nil, err
	}
	return rc, nil
}

package nvlist

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"testing"
)

func TestLooper(t *testing.T) {
	strOr := func(s ...string) string {
		for i := range s {
			if s[i] != "" {
				return s[i]
			}
		}
		return ""
	}
	path := strOr(os.Getenv("ZFSFILE"), "/obrovsky/recovery/zbackup0")
	if path == "" {
		t.Fatalf("path must be set.")
	}
	fh, err := os.Open(path)
	if err != nil {
		t.Fatalf("could not open %q; %v", path, err)
	}
	fh.Seek(0x4000, 0)
	nvlist := make([]byte, 0x1c000)
	if err := binary.Read(fh, binary.LittleEndian, nvlist); err != nil {
		t.Fatal(err)
	}

	// now that nvlist is a byte slice, we can make it a
	// bytes.Reader
	br := bytes.NewReader(nvlist)

	m, err := Read(br)

	if err != nil {
		t.Fatal(err)
	}

	o, err := json.MarshalIndent(m, "", "  ")

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("\n%s", string(o))
}

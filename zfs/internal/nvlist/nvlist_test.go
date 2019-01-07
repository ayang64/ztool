package nvlist_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/ayang64/ztool/internal/nvlist"
	"os"
	"testing"
)

func strOr(s ...string) string {
	for i := range s {
		if s[i] != "" {
			return s[i]
		}
	}
	return ""
}

func BenchmarkRead(b *testing.B) {

	path := strOr(os.Getenv("ZFSFILE"), "/obrovsky/recovery/zbackup0")
	if path == "" {
		b.Fatalf("path must be set.")
	}
	fh, err := os.Open(path)
	if err != nil {
		b.Fatalf("could not open %q; %v", path, err)
	}
	fh.Seek(0x4000, 0)
	nvl := make([]byte, 0x1c000)
	if err := binary.Read(fh, binary.LittleEndian, nvl); err != nil {
		b.Fatal(err)
	}

	br := bytes.NewReader(nvl)
	for i := 0; i < b.N; i++ {
		nvlist.Read(br)
		br.Reset(nvl)
	}
}

func TestLooper(t *testing.T) {
	path := strOr(os.Getenv("ZFSFILE"), "/obrovsky/recovery/zbackup0")
	if path == "" {
		t.Fatalf("path must be set.")
	}
	fh, err := os.Open(path)
	if err != nil {
		t.Fatalf("could not open %q; %v", path, err)
	}
	fh.Seek(0x4000, 0)
	nvl := make([]byte, 0x1c000)
	if err := binary.Read(fh, binary.LittleEndian, nvl); err != nil {
		t.Fatal(err)
	}

	// now that nvlist is a byte slice, we can make it a
	// bytes.Reader
	br := bytes.NewReader(nvl)

	m, err := nvlist.Read(br)

	if err != nil {
		t.Fatal(err)
	}

	o, err := json.MarshalIndent(m, "", "  ")

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("\n%s", string(o))
}

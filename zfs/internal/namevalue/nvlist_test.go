package nvlist

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"testing"
	"unicode"
)

func TestNvlist(t *testing.T) {
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
		t.FailNow()
	}

	t.Logf("testing with %s", path)

	fh, err := os.Open(path)

	if err != nil {
		t.Fatalf("could not open %q; %v", path, err)
		t.FailNow()
	}

	fh.Seek(0x4000, 0)

	nvlist := make([]byte, 0x1c000)
	if err := binary.Read(fh, binary.LittleEndian, nvlist); err != nil {
		t.Fatal(err)
		t.FailNow()
	}

	// now that nvlist is a byte slice, we can make it a
	// bytes.Reader

	br := bytes.NewReader(nvlist)

	header := Header{}
	if err := binary.Read(br, binary.LittleEndian, &header); err != nil {
		t.Fatal(err)
		t.FailNow()
	}

	t.Logf("header = %v", header)

	list := List{}
	if err := binary.Read(br, binary.LittleEndian, &list); err != nil {
		t.Fatal(err)
		t.FailNow()
	}

	t.Logf("header = %#v", list)

	chunk := func(stride int, bytes []byte) [][]byte {
		rc := [][]byte{}
		for i, end := 0, len(bytes)/stride; i < end; i++ {
			rc = append(rc, bytes[i*stride:(i+1)*stride])
		}
		return rc
	}

	hex := func(b []byte) string {
		rc := ""
		for i := range b {
			rc += fmt.Sprintf("%02x ", b[i])
			if (i+1)%4 == 0 {
				rc += "  "
			}
		}
		return rc
	}

	ascii := func(b []byte) string {
		rc := ""
		printable := func(b byte) bool {
			if r := rune(b); r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsPunct(r) || unicode.IsDigit(r)) {
				return true
			}
			return false
		}
		for i := range b {
			rc += func() string {
				if printable(b[i]) {
					return string(rune(b[i]))
				}
				return "."
			}()
		}
		return rc
	}

	for _, s := range chunk(16, nvlist) {
		t.Logf("%s  %s", hex(s), ascii(s))
	}

	DumpNvlist(bytes.NewReader(nvlist))
}

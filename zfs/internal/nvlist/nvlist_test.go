package nvlist

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
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
	}
	t.Logf("testing with %s", path)
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

	header := Header{}
	if err := binary.Read(br, binary.LittleEndian, &header); err != nil {
		t.Fatal(err)
	}

	t.Logf("header = %v", header)

	list := List{}
	if err := binary.Read(br, binary.LittleEndian, &list); err != nil {
		t.Fatal(err)
	}

	t.Logf("list = %#v", list)

	// read first pair

	// xdr data is encoded in Big Endian
	for {
		pair := Pair{}
		if err := binary.Read(br, binary.BigEndian, &pair); err != nil {
			t.Fatal(err)
		}

		t.Logf("pair = %#v", pair)

		align4 := func(i int32) int32 {
			// round i up to nearest 4 bytes.  this works by masking off the lower
			// two bits after adding three.
			//
			// adding three will potentially overflow into the non-masked bits.
			return (i + 3) & ^3
		}

		namesize := int32(0)
		if err := binary.Read(br, binary.BigEndian, &namesize); err != nil {
			t.Fatal(err)
		}
		t.Logf("namesize = %d", namesize)

		// read namesize amount of data
		namebytes := make([]byte, align4(namesize))
		if err := binary.Read(br, binary.BigEndian, namebytes); err != nil {
			t.Fatal(err)
		}

		t.Logf("name = %q (%#v)", string(namebytes), namebytes)

		// read type
		typ := Type(0)
		if err := binary.Read(br, binary.BigEndian, &typ); err != nil {
			t.Fatal(err)
		}
		t.Logf("type = %s (%d bytes)", typ, typ.Size())

		numelements := int32(0)
		if err := binary.Read(br, binary.BigEndian, &numelements); err != nil {
			t.Fatal(err)
		}

		t.Logf("numelements = %d", numelements)

		datasize := func() int32 {
			if numelements == 0 {
				return typ.Size()
			}
			return numelements * typ.Size()
		}()

		t.Logf("data size = %d", datasize)

		data := make([]byte, align4(datasize))
		if err := binary.Read(br, binary.BigEndian, data); err != nil {
			t.Fatal(err)
		}

		t.Logf("data = %#v", data)
	}

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

}

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
	t.Logf("testing with %s", path)
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

	/*
		scanner := NewScanner(br)

		t.Logf("s: %#v", scanner)

		for scanner.Next() {
			t.Logf("s: %s (%s), %d, %d -> %q", scanner.Name(), scanner.Type(), scanner.NumElements(), scanner.FieldSize(), scanner.ValueString())

		}

		if err := scanner.Error(); err != nil {
			t.Logf("errored field size: %d", scanner.FieldSize())
			t.Logf("     errored bytes: %#v", scanner.Bytes())
			t.Logf("      errored name: %s", scanner.Name())
			t.Fatal(err)
		}

	*/

	m, err := ReadFull(br)

	t.Logf("m = %#v", m)

	o, err := json.MarshalIndent(m, "", "  ")

	t.Logf("%s", string(o))

}

package nvlist_test

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/ayang64/ztool/zfs/internal/nvlist"
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
	fh, err := os.Open(path)
	if err != nil {
		t.Fatalf("could not open %q; %v -- must set ZFSFILE", path, err)
	}

	// seek to beginning of nvlist and clamp reader to its size.
	//
	// zfs nvlist is XDR encoded data that lives between 0x4000 - 0x20000 on the volume.
	nvr := io.NewSectionReader(fh, 0x4000, 0x1c000)

	m, err := nvlist.Read(nvr)

	if err != nil {
		t.Fatal(err)
	}

	o, err := json.MarshalIndent(m, "", "  ")

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("\n%s", string(o))
}

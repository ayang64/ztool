package zfs_test

import (
	"log"
	"reflect"
	"testing"

	"github.com/ayang64/ztool/zfs"
	_ "github.com/pierrec/lz4"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// This is the big test.
func TestZFS(t *testing.T) {
	tests := []struct {
		Name string
		Path string
	}{
		{Name: "simple 100mb file", Path: "../test-data/uncompressed-simple.image"},
		{Name: "original disk", Path: "../../zbackup-editable"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			// get a new filesystem.
			fs, err := zfs.New(zfs.WithPath(test.Path))

			if err != nil {
				t.Fatal(err)
			}

			ashift, _ := fs.AShift()

			t.Logf("ashift = %d", ashift)

			ubs, err := fs.UberBlocks()
			ubs = ubs
			if err != nil {
				t.Fatal(err)
			}

			ub, err := fs.ActiveUberBlock()

			if err != nil {
				t.Fatal(err)
			}

			for idx, vd := range ub.RootBP.Vdevs {
				t.Logf("compression: %s", ub.RootBP.Props.Compression())
				t.Logf("vdev %03d (%d): %#v, gang = %v, disk offset = %d", vd.VDEV, idx, vd, vd.Gang(), vd.Block())
				t.Logf("    Asize: %d, Size: %d", vd.Asize(), vd.Size)
			}
			t.Logf("ROOTBP: %s", ub.RootBP)

			t.Logf("ub.RootBP.Props.Psize() = %d", ub.RootBP.Props.Psize())
			t.Logf("ub.RootBP.Props.Lsize() = %d", ub.RootBP.Props.Lsize())
			t.Logf("ub.RootBP.Props.Type() = %d", ub.RootBP.Props.Type())

			t.Logf("%d/%d", ub.RootBP.Vdevs[0].Block(), ub.RootBP.Vdevs[0].Offset)

			rb, err := fs.GetDnode(&ub.RootBP)

			if err != nil {
				t.Fatal(err)
			}
			t.Logf("ROOT BLOCK/MOS: %v", rb)
		})
	}
}

func TestSizeofs(t *testing.T) {
	tests := []struct {
		Name         string
		Value        interface{}
		ExpectedSize uintptr
	}{
		{Name: "BlockPointer", Value: zfs.BlockPointer{}, ExpectedSize: 128},
		{Name: "DnodePhys", Value: zfs.DnodePhys{}, ExpectedSize: 512},
		{Name: "UberBlock", Value: zfs.UberBlock{}, ExpectedSize: 208},
		{Name: "VdevLabel", Value: zfs.VdevLabel{}, ExpectedSize: 262144},
		{Name: "DVA", Value: zfs.DVA{}, ExpectedSize: 16},
	}

	t.Parallel()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if typ := reflect.TypeOf(test.Value); typ.Size() != test.ExpectedSize {
				t.Fatalf("Size of %s is %d bytes; expected %d",
					typ.Name(), typ.Size(), test.ExpectedSize)
				t.FailNow()
			}
		})
	}
}

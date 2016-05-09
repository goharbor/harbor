package azure

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	azure "github.com/Azure/azure-sdk-for-go/storage"
)

func TestRandomWriter_writeChunkToBlocks(t *testing.T) {
	s := NewStorageSimulator()
	rw := newRandomBlobWriter(&s, 3)
	rand := newBlockIDGenerator()
	c := []byte("AAABBBCCCD")

	if err := rw.bs.CreateBlockBlob("a", "b"); err != nil {
		t.Fatal(err)
	}
	bw, nn, err := rw.writeChunkToBlocks("a", "b", bytes.NewReader(c), rand)
	if err != nil {
		t.Fatal(err)
	}
	if expected := int64(len(c)); nn != expected {
		t.Fatalf("wrong nn:%v, expected:%v", nn, expected)
	}
	if expected := 4; len(bw) != expected {
		t.Fatal("unexpected written block count")
	}

	bx, err := s.GetBlockList("a", "b", azure.BlockListTypeAll)
	if err != nil {
		t.Fatal(err)
	}
	if expected := 0; len(bx.CommittedBlocks) != expected {
		t.Fatal("unexpected committed block count")
	}
	if expected := 4; len(bx.UncommittedBlocks) != expected {
		t.Fatalf("unexpected uncommitted block count: %d -- %#v", len(bx.UncommittedBlocks), bx)
	}

	if err := rw.bs.PutBlockList("a", "b", bw); err != nil {
		t.Fatal(err)
	}

	r, err := rw.bs.GetBlob("a", "b")
	if err != nil {
		t.Fatal(err)
	}
	assertBlobContents(t, r, c)
}

func TestRandomWriter_blocksLeftSide(t *testing.T) {
	blob := "AAAAABBBBBCCC"
	cases := []struct {
		offset          int64
		expectedBlob    string
		expectedPattern []azure.BlockStatus
	}{
		{0, "", []azure.BlockStatus{}},                                                                                      // write to beginning, discard all
		{13, blob, []azure.BlockStatus{azure.BlockStatusCommitted, azure.BlockStatusCommitted, azure.BlockStatusCommitted}}, // write to end, no change
		{1, "A", []azure.BlockStatus{azure.BlockStatusUncommitted}},                                                         // write at 1
		{5, "AAAAA", []azure.BlockStatus{azure.BlockStatusCommitted}},                                                       // write just after first block
		{6, "AAAAAB", []azure.BlockStatus{azure.BlockStatusCommitted, azure.BlockStatusUncommitted}},                        // split the second block
		{9, "AAAAABBBB", []azure.BlockStatus{azure.BlockStatusCommitted, azure.BlockStatusUncommitted}},                     // write just after first block
	}

	for _, c := range cases {
		s := NewStorageSimulator()
		rw := newRandomBlobWriter(&s, 5)
		rand := newBlockIDGenerator()

		if err := rw.bs.CreateBlockBlob("a", "b"); err != nil {
			t.Fatal(err)
		}
		bw, _, err := rw.writeChunkToBlocks("a", "b", strings.NewReader(blob), rand)
		if err != nil {
			t.Fatal(err)
		}
		if err := rw.bs.PutBlockList("a", "b", bw); err != nil {
			t.Fatal(err)
		}
		bx, err := rw.blocksLeftSide("a", "b", c.offset, rand)
		if err != nil {
			t.Fatal(err)
		}

		bs := []azure.BlockStatus{}
		for _, v := range bx {
			bs = append(bs, v.Status)
		}

		if !reflect.DeepEqual(bs, c.expectedPattern) {
			t.Logf("Committed blocks %v", bw)
			t.Fatalf("For offset %v: Expected pattern: %v, Got: %v\n(Returned: %v)", c.offset, c.expectedPattern, bs, bx)
		}
		if rw.bs.PutBlockList("a", "b", bx); err != nil {
			t.Fatal(err)
		}
		r, err := rw.bs.GetBlob("a", "b")
		if err != nil {
			t.Fatal(err)
		}
		cout, err := ioutil.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		outBlob := string(cout)
		if outBlob != c.expectedBlob {
			t.Fatalf("wrong blob contents: %v, expected: %v", outBlob, c.expectedBlob)
		}
	}
}

func TestRandomWriter_blocksRightSide(t *testing.T) {
	blob := "AAAAABBBBBCCC"
	cases := []struct {
		offset          int64
		size            int64
		expectedBlob    string
		expectedPattern []azure.BlockStatus
	}{
		{0, 100, "", []azure.BlockStatus{}},                                                                                             // overwrite the entire blob
		{0, 3, "AABBBBBCCC", []azure.BlockStatus{azure.BlockStatusUncommitted, azure.BlockStatusCommitted, azure.BlockStatusCommitted}}, // split first block
		{4, 1, "BBBBBCCC", []azure.BlockStatus{azure.BlockStatusCommitted, azure.BlockStatusCommitted}},                                 // write to last char of first block
		{1, 6, "BBBCCC", []azure.BlockStatus{azure.BlockStatusUncommitted, azure.BlockStatusCommitted}},                                 // overwrite splits first and second block, last block remains
		{3, 8, "CC", []azure.BlockStatus{azure.BlockStatusUncommitted}},                                                                 // overwrite a block in middle block, split end block
		{10, 1, "CC", []azure.BlockStatus{azure.BlockStatusUncommitted}},                                                                // overwrite first byte of rightmost block
		{11, 2, "", []azure.BlockStatus{}},                                                                                              // overwrite the rightmost index
		{13, 20, "", []azure.BlockStatus{}},                                                                                             // append to the end
	}

	for _, c := range cases {
		s := NewStorageSimulator()
		rw := newRandomBlobWriter(&s, 5)
		rand := newBlockIDGenerator()

		if err := rw.bs.CreateBlockBlob("a", "b"); err != nil {
			t.Fatal(err)
		}
		bw, _, err := rw.writeChunkToBlocks("a", "b", strings.NewReader(blob), rand)
		if err != nil {
			t.Fatal(err)
		}
		if err := rw.bs.PutBlockList("a", "b", bw); err != nil {
			t.Fatal(err)
		}
		bx, err := rw.blocksRightSide("a", "b", c.offset, c.size, rand)
		if err != nil {
			t.Fatal(err)
		}

		bs := []azure.BlockStatus{}
		for _, v := range bx {
			bs = append(bs, v.Status)
		}

		if !reflect.DeepEqual(bs, c.expectedPattern) {
			t.Logf("Committed blocks %v", bw)
			t.Fatalf("For offset %v-size:%v: Expected pattern: %v, Got: %v\n(Returned: %v)", c.offset, c.size, c.expectedPattern, bs, bx)
		}
		if rw.bs.PutBlockList("a", "b", bx); err != nil {
			t.Fatal(err)
		}
		r, err := rw.bs.GetBlob("a", "b")
		if err != nil {
			t.Fatal(err)
		}
		cout, err := ioutil.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		outBlob := string(cout)
		if outBlob != c.expectedBlob {
			t.Fatalf("For offset %v-size:%v: wrong blob contents: %v, expected: %v", c.offset, c.size, outBlob, c.expectedBlob)
		}
	}
}

func TestRandomWriter_Write_NewBlob(t *testing.T) {
	var (
		s    = NewStorageSimulator()
		rw   = newRandomBlobWriter(&s, 1024*3) // 3 KB blocks
		blob = randomContents(1024 * 7)        // 7 KB blob
	)
	if err := rw.bs.CreateBlockBlob("a", "b"); err != nil {
		t.Fatal(err)
	}

	if _, err := rw.WriteBlobAt("a", "b", 10, bytes.NewReader(blob)); err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, err := rw.WriteBlobAt("a", "b", 100000, bytes.NewReader(blob)); err == nil {
		t.Fatal("expected error, got nil")
	}
	if nn, err := rw.WriteBlobAt("a", "b", 0, bytes.NewReader(blob)); err != nil {
		t.Fatal(err)
	} else if expected := int64(len(blob)); expected != nn {
		t.Fatalf("wrong written bytes count: %v, expected: %v", nn, expected)
	}
	if out, err := rw.bs.GetBlob("a", "b"); err != nil {
		t.Fatal(err)
	} else {
		assertBlobContents(t, out, blob)
	}
	if bx, err := rw.bs.GetBlockList("a", "b", azure.BlockListTypeCommitted); err != nil {
		t.Fatal(err)
	} else if len(bx.CommittedBlocks) != 3 {
		t.Fatalf("got wrong number of committed blocks: %v", len(bx.CommittedBlocks))
	}

	// Replace first 512 bytes
	leftChunk := randomContents(512)
	blob = append(leftChunk, blob[512:]...)
	if nn, err := rw.WriteBlobAt("a", "b", 0, bytes.NewReader(leftChunk)); err != nil {
		t.Fatal(err)
	} else if expected := int64(len(leftChunk)); expected != nn {
		t.Fatalf("wrong written bytes count: %v, expected: %v", nn, expected)
	}
	if out, err := rw.bs.GetBlob("a", "b"); err != nil {
		t.Fatal(err)
	} else {
		assertBlobContents(t, out, blob)
	}
	if bx, err := rw.bs.GetBlockList("a", "b", azure.BlockListTypeCommitted); err != nil {
		t.Fatal(err)
	} else if expected := 4; len(bx.CommittedBlocks) != expected {
		t.Fatalf("got wrong number of committed blocks: %v, expected: %v", len(bx.CommittedBlocks), expected)
	}

	// Replace last 512 bytes with 1024 bytes
	rightChunk := randomContents(1024)
	offset := int64(len(blob) - 512)
	blob = append(blob[:offset], rightChunk...)
	if nn, err := rw.WriteBlobAt("a", "b", offset, bytes.NewReader(rightChunk)); err != nil {
		t.Fatal(err)
	} else if expected := int64(len(rightChunk)); expected != nn {
		t.Fatalf("wrong written bytes count: %v, expected: %v", nn, expected)
	}
	if out, err := rw.bs.GetBlob("a", "b"); err != nil {
		t.Fatal(err)
	} else {
		assertBlobContents(t, out, blob)
	}
	if bx, err := rw.bs.GetBlockList("a", "b", azure.BlockListTypeCommitted); err != nil {
		t.Fatal(err)
	} else if expected := 5; len(bx.CommittedBlocks) != expected {
		t.Fatalf("got wrong number of committed blocks: %v, expected: %v", len(bx.CommittedBlocks), expected)
	}

	// Replace 2K-4K (overlaps 2 blocks from L/R)
	newChunk := randomContents(1024 * 2)
	offset = 1024 * 2
	blob = append(append(blob[:offset], newChunk...), blob[offset+int64(len(newChunk)):]...)
	if nn, err := rw.WriteBlobAt("a", "b", offset, bytes.NewReader(newChunk)); err != nil {
		t.Fatal(err)
	} else if expected := int64(len(newChunk)); expected != nn {
		t.Fatalf("wrong written bytes count: %v, expected: %v", nn, expected)
	}
	if out, err := rw.bs.GetBlob("a", "b"); err != nil {
		t.Fatal(err)
	} else {
		assertBlobContents(t, out, blob)
	}
	if bx, err := rw.bs.GetBlockList("a", "b", azure.BlockListTypeCommitted); err != nil {
		t.Fatal(err)
	} else if expected := 6; len(bx.CommittedBlocks) != expected {
		t.Fatalf("got wrong number of committed blocks: %v, expected: %v\n%v", len(bx.CommittedBlocks), expected, bx.CommittedBlocks)
	}

	// Replace the entire blob
	newBlob := randomContents(1024 * 30)
	if nn, err := rw.WriteBlobAt("a", "b", 0, bytes.NewReader(newBlob)); err != nil {
		t.Fatal(err)
	} else if expected := int64(len(newBlob)); expected != nn {
		t.Fatalf("wrong written bytes count: %v, expected: %v", nn, expected)
	}
	if out, err := rw.bs.GetBlob("a", "b"); err != nil {
		t.Fatal(err)
	} else {
		assertBlobContents(t, out, newBlob)
	}
	if bx, err := rw.bs.GetBlockList("a", "b", azure.BlockListTypeCommitted); err != nil {
		t.Fatal(err)
	} else if expected := 10; len(bx.CommittedBlocks) != expected {
		t.Fatalf("got wrong number of committed blocks: %v, expected: %v\n%v", len(bx.CommittedBlocks), expected, bx.CommittedBlocks)
	} else if expected, size := int64(1024*30), getBlobSize(bx); size != expected {
		t.Fatalf("committed block size does not indicate blob size")
	}
}

func Test_getBlobSize(t *testing.T) {
	// with some committed blocks
	if expected, size := int64(151), getBlobSize(azure.BlockListResponse{
		CommittedBlocks: []azure.BlockResponse{
			{"A", 100},
			{"B", 50},
			{"C", 1},
		},
		UncommittedBlocks: []azure.BlockResponse{
			{"D", 200},
		}}); expected != size {
		t.Fatalf("wrong blob size: %v, expected: %v", size, expected)
	}

	// with no committed blocks
	if expected, size := int64(0), getBlobSize(azure.BlockListResponse{
		UncommittedBlocks: []azure.BlockResponse{
			{"A", 100},
			{"B", 50},
			{"C", 1},
			{"D", 200},
		}}); expected != size {
		t.Fatalf("wrong blob size: %v, expected: %v", size, expected)
	}
}

func assertBlobContents(t *testing.T, r io.Reader, expected []byte) {
	out, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("wrong blob contents. size: %v, expected: %v", len(out), len(expected))
	}
}

func randomContents(length int64) []byte {
	b := make([]byte, length)
	for i := range b {
		b[i] = byte(rand.Intn(2 << 8))
	}
	return b
}

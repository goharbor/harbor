package azure

import (
	"bytes"
	"testing"
)

func Test_zeroFillWrite_AppendNoGap(t *testing.T) {
	s := NewStorageSimulator()
	bw := newRandomBlobWriter(&s, 1024*1)
	zw := newZeroFillWriter(&bw)
	if err := s.CreateBlockBlob("a", "b"); err != nil {
		t.Fatal(err)
	}

	firstChunk := randomContents(1024*3 + 512)
	if nn, err := zw.Write("a", "b", 0, bytes.NewReader(firstChunk)); err != nil {
		t.Fatal(err)
	} else if expected := int64(len(firstChunk)); expected != nn {
		t.Fatalf("wrong written bytes count: %v, expected: %v", nn, expected)
	}
	if out, err := s.GetBlob("a", "b"); err != nil {
		t.Fatal(err)
	} else {
		assertBlobContents(t, out, firstChunk)
	}

	secondChunk := randomContents(256)
	if nn, err := zw.Write("a", "b", int64(len(firstChunk)), bytes.NewReader(secondChunk)); err != nil {
		t.Fatal(err)
	} else if expected := int64(len(secondChunk)); expected != nn {
		t.Fatalf("wrong written bytes count: %v, expected: %v", nn, expected)
	}
	if out, err := s.GetBlob("a", "b"); err != nil {
		t.Fatal(err)
	} else {
		assertBlobContents(t, out, append(firstChunk, secondChunk...))
	}

}

func Test_zeroFillWrite_StartWithGap(t *testing.T) {
	s := NewStorageSimulator()
	bw := newRandomBlobWriter(&s, 1024*2)
	zw := newZeroFillWriter(&bw)
	if err := s.CreateBlockBlob("a", "b"); err != nil {
		t.Fatal(err)
	}

	chunk := randomContents(1024 * 5)
	padding := int64(1024*2 + 256)
	if nn, err := zw.Write("a", "b", padding, bytes.NewReader(chunk)); err != nil {
		t.Fatal(err)
	} else if expected := int64(len(chunk)); expected != nn {
		t.Fatalf("wrong written bytes count: %v, expected: %v", nn, expected)
	}
	if out, err := s.GetBlob("a", "b"); err != nil {
		t.Fatal(err)
	} else {
		assertBlobContents(t, out, append(make([]byte, padding), chunk...))
	}
}

func Test_zeroFillWrite_AppendWithGap(t *testing.T) {
	s := NewStorageSimulator()
	bw := newRandomBlobWriter(&s, 1024*2)
	zw := newZeroFillWriter(&bw)
	if err := s.CreateBlockBlob("a", "b"); err != nil {
		t.Fatal(err)
	}

	firstChunk := randomContents(1024*3 + 512)
	if _, err := zw.Write("a", "b", 0, bytes.NewReader(firstChunk)); err != nil {
		t.Fatal(err)
	}
	if out, err := s.GetBlob("a", "b"); err != nil {
		t.Fatal(err)
	} else {
		assertBlobContents(t, out, firstChunk)
	}

	secondChunk := randomContents(256)
	padding := int64(1024 * 4)
	if nn, err := zw.Write("a", "b", int64(len(firstChunk))+padding, bytes.NewReader(secondChunk)); err != nil {
		t.Fatal(err)
	} else if expected := int64(len(secondChunk)); expected != nn {
		t.Fatalf("wrong written bytes count: %v, expected: %v", nn, expected)
	}
	if out, err := s.GetBlob("a", "b"); err != nil {
		t.Fatal(err)
	} else {
		assertBlobContents(t, out, append(firstChunk, append(make([]byte, padding), secondChunk...)...))
	}
}

func Test_zeroFillWrite_LiesWithinSize(t *testing.T) {
	s := NewStorageSimulator()
	bw := newRandomBlobWriter(&s, 1024*2)
	zw := newZeroFillWriter(&bw)
	if err := s.CreateBlockBlob("a", "b"); err != nil {
		t.Fatal(err)
	}

	firstChunk := randomContents(1024 * 3)
	if _, err := zw.Write("a", "b", 0, bytes.NewReader(firstChunk)); err != nil {
		t.Fatal(err)
	}
	if out, err := s.GetBlob("a", "b"); err != nil {
		t.Fatal(err)
	} else {
		assertBlobContents(t, out, firstChunk)
	}

	// in this case, zerofill won't be used
	secondChunk := randomContents(256)
	if nn, err := zw.Write("a", "b", 0, bytes.NewReader(secondChunk)); err != nil {
		t.Fatal(err)
	} else if expected := int64(len(secondChunk)); expected != nn {
		t.Fatalf("wrong written bytes count: %v, expected: %v", nn, expected)
	}
	if out, err := s.GetBlob("a", "b"); err != nil {
		t.Fatal(err)
	} else {
		assertBlobContents(t, out, append(secondChunk, firstChunk[len(secondChunk):]...))
	}
}

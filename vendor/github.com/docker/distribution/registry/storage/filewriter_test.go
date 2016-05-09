package storage

import (
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"testing"

	"github.com/docker/distribution/context"
	"github.com/docker/distribution/digest"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/inmemory"
)

// TestSimpleWrite takes the fileWriter through common write operations
// ensuring data integrity.
func TestSimpleWrite(t *testing.T) {
	content := make([]byte, 1<<20)
	n, err := rand.Read(content)
	if err != nil {
		t.Fatalf("unexpected error building random data: %v", err)
	}

	if n != len(content) {
		t.Fatalf("random read did't fill buffer")
	}

	dgst, err := digest.FromReader(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error digesting random content: %v", err)
	}

	driver := inmemory.New()
	path := "/random"
	ctx := context.Background()

	fw, err := newFileWriter(ctx, driver, path)
	if err != nil {
		t.Fatalf("unexpected error creating fileWriter: %v", err)
	}
	defer fw.Close()

	n, err = fw.Write(content)
	if err != nil {
		t.Fatalf("unexpected error writing content: %v", err)
	}

	if n != len(content) {
		t.Fatalf("unexpected write length: %d != %d", n, len(content))
	}

	fr, err := newFileReader(ctx, driver, path, int64(len(content)))
	if err != nil {
		t.Fatalf("unexpected error creating fileReader: %v", err)
	}
	defer fr.Close()

	verifier, err := digest.NewDigestVerifier(dgst)
	if err != nil {
		t.Fatalf("unexpected error getting digest verifier: %s", err)
	}

	io.Copy(verifier, fr)

	if !verifier.Verified() {
		t.Fatalf("unable to verify write data")
	}

	// Check the seek position is equal to the content length
	end, err := fw.Seek(0, os.SEEK_END)
	if err != nil {
		t.Fatalf("unexpected error seeking: %v", err)
	}

	if end != int64(len(content)) {
		t.Fatalf("write did not advance offset: %d != %d", end, len(content))
	}

	// Double the content
	doubled := append(content, content...)
	doubledgst, err := digest.FromReader(bytes.NewReader(doubled))
	if err != nil {
		t.Fatalf("unexpected error digesting doubled content: %v", err)
	}

	nn, err := fw.ReadFrom(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error doubling content: %v", err)
	}

	if nn != int64(len(content)) {
		t.Fatalf("writeat was short: %d != %d", n, len(content))
	}

	fr, err = newFileReader(ctx, driver, path, int64(len(doubled)))
	if err != nil {
		t.Fatalf("unexpected error creating fileReader: %v", err)
	}
	defer fr.Close()

	verifier, err = digest.NewDigestVerifier(doubledgst)
	if err != nil {
		t.Fatalf("unexpected error getting digest verifier: %s", err)
	}

	io.Copy(verifier, fr)

	if !verifier.Verified() {
		t.Fatalf("unable to verify write data")
	}

	// Check that Write updated the offset.
	end, err = fw.Seek(0, os.SEEK_END)
	if err != nil {
		t.Fatalf("unexpected error seeking: %v", err)
	}

	if end != int64(len(doubled)) {
		t.Fatalf("write did not advance offset: %d != %d", end, len(doubled))
	}

	// Now, we copy from one path to another, running the data through the
	// fileReader to fileWriter, rather than the driver.Move command to ensure
	// everything is working correctly.
	fr, err = newFileReader(ctx, driver, path, int64(len(doubled)))
	if err != nil {
		t.Fatalf("unexpected error creating fileReader: %v", err)
	}
	defer fr.Close()

	fw, err = newFileWriter(ctx, driver, "/copied")
	if err != nil {
		t.Fatalf("unexpected error creating fileWriter: %v", err)
	}
	defer fw.Close()

	nn, err = io.Copy(fw, fr)
	if err != nil {
		t.Fatalf("unexpected error copying data: %v", err)
	}

	if nn != int64(len(doubled)) {
		t.Fatalf("unexpected copy length: %d != %d", nn, len(doubled))
	}

	fr, err = newFileReader(ctx, driver, "/copied", int64(len(doubled)))
	if err != nil {
		t.Fatalf("unexpected error creating fileReader: %v", err)
	}
	defer fr.Close()

	verifier, err = digest.NewDigestVerifier(doubledgst)
	if err != nil {
		t.Fatalf("unexpected error getting digest verifier: %s", err)
	}

	io.Copy(verifier, fr)

	if !verifier.Verified() {
		t.Fatalf("unable to verify write data")
	}
}

func BenchmarkFileWriter(b *testing.B) {
	b.StopTimer() // not sure how long setup above will take
	for i := 0; i < b.N; i++ {
		// Start basic fileWriter initialization
		fw := fileWriter{
			driver: inmemory.New(),
			path:   "/random",
		}
		ctx := context.Background()
		if fi, err := fw.driver.Stat(ctx, fw.path); err != nil {
			switch err := err.(type) {
			case storagedriver.PathNotFoundError:
				// ignore, offset is zero
			default:
				b.Fatalf("Failed to initialize fileWriter: %v", err.Error())
			}
		} else {
			if fi.IsDir() {
				b.Fatalf("Cannot write to a directory")
			}

			fw.size = fi.Size()
		}

		randomBytes := make([]byte, 1<<20)
		_, err := rand.Read(randomBytes)
		if err != nil {
			b.Fatalf("unexpected error building random data: %v", err)
		}
		// End basic file writer initialization

		b.StartTimer()
		for j := 0; j < 100; j++ {
			fw.Write(randomBytes)
		}
		b.StopTimer()
	}
}

func BenchmarkfileWriter(b *testing.B) {
	b.StopTimer() // not sure how long setup above will take
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		bfw, err := newFileWriter(ctx, inmemory.New(), "/random")

		if err != nil {
			b.Fatalf("Failed to initialize fileWriter: %v", err.Error())
		}

		randomBytes := make([]byte, 1<<20)
		_, err = rand.Read(randomBytes)
		if err != nil {
			b.Fatalf("unexpected error building random data: %v", err)
		}

		b.StartTimer()
		for j := 0; j < 100; j++ {
			bfw.Write(randomBytes)
		}
		b.StopTimer()
	}
}

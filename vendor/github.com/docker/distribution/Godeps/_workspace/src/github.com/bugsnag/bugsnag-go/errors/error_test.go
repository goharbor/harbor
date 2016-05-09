package errors

import (
	"bytes"
	"fmt"
	"io"
	"runtime/debug"
	"testing"
)

func TestStackFormatMatches(t *testing.T) {

	defer func() {
		err := recover()
		if err != 'a' {
			t.Fatal(err)
		}

		bs := [][]byte{Errorf("hi").Stack(), debug.Stack()}

		// Ignore the first line (as it contains the PC of the .Stack() call)
		bs[0] = bytes.SplitN(bs[0], []byte("\n"), 2)[1]
		bs[1] = bytes.SplitN(bs[1], []byte("\n"), 2)[1]

		if bytes.Compare(bs[0], bs[1]) != 0 {
			t.Errorf("Stack didn't match")
			t.Errorf("%s", bs[0])
			t.Errorf("%s", bs[1])
		}
	}()

	a()
}

func TestSkipWorks(t *testing.T) {

	defer func() {
		err := recover()
		if err != 'a' {
			t.Fatal(err)
		}

		bs := [][]byte{New("hi", 2).Stack(), debug.Stack()}

		// should skip four lines of debug.Stack()
		bs[1] = bytes.SplitN(bs[1], []byte("\n"), 5)[4]

		if bytes.Compare(bs[0], bs[1]) != 0 {
			t.Errorf("Stack didn't match")
			t.Errorf("%s", bs[0])
			t.Errorf("%s", bs[1])
		}
	}()

	a()
}

func TestNewError(t *testing.T) {

	e := func() error {
		return New("hi", 1)
	}()

	if e.Error() != "hi" {
		t.Errorf("Constructor with a string failed")
	}

	if New(fmt.Errorf("yo"), 0).Error() != "yo" {
		t.Errorf("Constructor with an error failed")
	}

	if New(e, 0) != e {
		t.Errorf("Constructor with an Error failed")
	}

	if New(nil, 0).Error() != "<nil>" {
		t.Errorf("Constructor with nil failed")
	}
}

func ExampleErrorf(x int) (int, error) {
	if x%2 == 1 {
		return 0, Errorf("can only halve even numbers, got %d", x)
	}
	return x / 2, nil
}

func ExampleNewError() (error, error) {
	// Wrap io.EOF with the current stack-trace and return it
	return nil, New(io.EOF, 0)
}

func ExampleNewError_skip() {
	defer func() {
		if err := recover(); err != nil {
			// skip 1 frame (the deferred function) and then return the wrapped err
			err = New(err, 1)
		}
	}()
}

func ExampleError_Stack(err Error) {
	fmt.Printf("Error: %s\n%s", err.Error(), err.Stack())
}

func a() error {
	b(5)
	return nil
}

func b(i int) {
	c()
}

func c() {
	panic('a')
}

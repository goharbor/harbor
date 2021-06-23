// Package parth provides path parsing for segment unmarshaling and slicing. In
// other words, parth provides simple and flexible access to (URL) path
// parameters.
//
// Along with string, all basic non-alias types are supported. An interface is
// available for implementation by user-defined types. When handling an int,
// uint, or float of any size, the first valid value within the specified
// segment will be used.
package parth

import (
	"errors"
)

// Unmarshaler is the interface implemented by types that can unmarshal a path
// segment representation of themselves. It is safe to assume that the segment
// data will not include slashes.
type Unmarshaler interface {
	UnmarshalSegment(string) error
}

// Err{Name} values facilitate error identification.
var (
	ErrUnknownType = errors.New("unknown type provided")

	ErrFirstSegNotFound = errors.New("first segment not found by index")
	ErrLastSegNotFound  = errors.New("last segment not found by index")
	ErrSegOrderReversed = errors.New("first segment must precede last segment")
	ErrKeySegNotFound   = errors.New("segment not found by key")

	ErrDataUnparsable = errors.New("data cannot be parsed")
)

// Segment locates the path segment indicated by the index i and unmarshals it
// into the provided type v. If the index is negative, the negative count
// begins with the last segment. An error is returned if: 1. The type is not a
// pointer to an instance of one of the basic non-alias types and does not
// implement the Unmarshaler interface; 2. The index is out of range of the
// path; 3. The located path segment data cannot be parsed as the provided type
// or if an error is returned when using a provided Unmarshaler implementation.
func Segment(path string, i int, v interface{}) error { //nolint
	var err error

	switch v := v.(type) {
	case *bool:
		*v, err = segmentToBool(path, i)

	case *float32:
		var f float64
		f, err = segmentToFloatN(path, i, 32)
		*v = float32(f)

	case *float64:
		*v, err = segmentToFloatN(path, i, 64)

	case *int:
		var n int64
		n, err = segmentToIntN(path, i, 0)
		*v = int(n)

	case *int16:
		var n int64
		n, err = segmentToIntN(path, i, 16)
		*v = int16(n)

	case *int32:
		var n int64
		n, err = segmentToIntN(path, i, 32)
		*v = int32(n)

	case *int64:
		*v, err = segmentToIntN(path, i, 64)

	case *int8:
		var n int64
		n, err = segmentToIntN(path, i, 8)
		*v = int8(n)

	case *string:
		*v, err = segmentToString(path, i)

	case *uint:
		var n uint64
		n, err = segmentToUintN(path, i, 0)
		*v = uint(n)

	case *uint16:
		var n uint64
		n, err = segmentToUintN(path, i, 16)
		*v = uint16(n)

	case *uint32:
		var n uint64
		n, err = segmentToUintN(path, i, 32)
		*v = uint32(n)

	case *uint64:
		*v, err = segmentToUintN(path, i, 64)

	case *uint8:
		var n uint64
		n, err = segmentToUintN(path, i, 8)
		*v = uint8(n)

	case Unmarshaler:
		var s string
		s, err = segmentToString(path, i)
		if err == nil {
			err = v.UnmarshalSegment(s)
		}

	default:
		err = ErrUnknownType
	}

	return err
}

// Sequent is similar to Segment, but uses a key to locate a segment and then
// unmarshal the subsequent segment. It is a simple wrapper over SubSeg with an
// index of 0.
func Sequent(path, key string, v interface{}) error {
	return SubSeg(path, key, 0, v)
}

// Span returns the path segments between two segment indexes i and j including
// the first segment. If an index is negative, the negative count begins with
// the last segment. Providing a 0 for the last index j is a special case which
// acts as an alias for the end of the path. If the first segment does not begin
// with a slash and it is part of the requested span, no slash will be added. An
// error is returned if: 1. Either index is out of range of the path; 2. The
// first index i does not precede the last index j.
func Span(path string, i, j int) (string, error) {
	var f, l int
	var ok bool

	if i < 0 {
		f, ok = segStartIndexFromEnd(path, i)
	} else {
		f, ok = segStartIndexFromStart(path, i)
	}
	if !ok {
		return "", ErrFirstSegNotFound
	}

	if j > 0 {
		l, ok = segEndIndexFromStart(path, j)
	} else {
		l, ok = segEndIndexFromEnd(path, j)
	}
	if !ok {
		return "", ErrLastSegNotFound
	}

	if f == l {
		return "", nil
	}

	if f > l {
		return "", ErrSegOrderReversed
	}

	return path[f:l], nil
}

// SubSeg is similar to Segment, but only handles the portion of the path
// subsequent to the provided key. For example, to access the segment
// immediately after a key, an index of 0 should be provided (see Sequent). An
// error is returned if the key cannot be found in the path.
func SubSeg(path, key string, i int, v interface{}) error { //nolint
	var err error

	switch v := v.(type) {
	case *bool:
		*v, err = subSegToBool(path, key, i)

	case *float32:
		var f float64
		f, err = subSegToFloatN(path, key, i, 32)
		*v = float32(f)

	case *float64:
		*v, err = subSegToFloatN(path, key, i, 64)

	case *int:
		var n int64
		n, err = subSegToIntN(path, key, i, 0)
		*v = int(n)

	case *int16:
		var n int64
		n, err = subSegToIntN(path, key, i, 16)
		*v = int16(n)

	case *int32:
		var n int64
		n, err = subSegToIntN(path, key, i, 32)
		*v = int32(n)

	case *int64:
		*v, err = subSegToIntN(path, key, i, 64)

	case *int8:
		var n int64
		n, err = subSegToIntN(path, key, i, 8)
		*v = int8(n)

	case *string:
		*v, err = subSegToString(path, key, i)

	case *uint:
		var n uint64
		n, err = subSegToUintN(path, key, i, 0)
		*v = uint(n)

	case *uint16:
		var n uint64
		n, err = subSegToUintN(path, key, i, 16)
		*v = uint16(n)

	case *uint32:
		var n uint64
		n, err = subSegToUintN(path, key, i, 32)
		*v = uint32(n)

	case *uint64:
		*v, err = subSegToUintN(path, key, i, 64)

	case *uint8:
		var n uint64
		n, err = subSegToUintN(path, key, i, 8)
		*v = uint8(n)

	case Unmarshaler:
		var s string
		s, err = subSegToString(path, key, i)
		if err == nil {
			err = v.UnmarshalSegment(s)
		}

	default:
		err = ErrUnknownType
	}

	return err
}

// SubSpan is similar to Span, but only handles the portion of the path
// subsequent to the provided key. An error is returned if the key cannot be
// found in the path.
func SubSpan(path, key string, i, j int) (string, error) {
	si, ok := segIndexByKey(path, key)
	if !ok {
		return "", ErrKeySegNotFound
	}

	if i >= 0 {
		i++
	}
	if j > 0 {
		j++
	}

	s, err := Span(path[si:], i, j)
	if err != nil {
		return "", err
	}

	return s, nil
}

// Parth manages path and error data for processing a single path multiple
// times while error checking only once. Only the first encountered error is
// stored as all subsequent calls to Parth methods that can error are elided.
type Parth struct {
	path string
	err  error
}

// New constructs a pointer to an instance of Parth around the provided path.
func New(path string) *Parth {
	return &Parth{path: path}
}

// NewBySpan constructs a pointer to an instance of Parth after preprocessing
// the provided path with Span.
func NewBySpan(path string, i, j int) *Parth {
	s, err := Span(path, i, j)
	return &Parth{s, err}
}

// NewBySubSpan constructs a pointer to an instance of Parth after
// preprocessing the provided path with SubSpan.
func NewBySubSpan(path, key string, i, j int) *Parth {
	s, err := SubSpan(path, key, i, j)
	return &Parth{s, err}
}

// Err returns the first error encountered by the *Parth receiver.
func (p *Parth) Err() error {
	return p.err
}

// Segment operates the same as the package-level function Segment.
func (p *Parth) Segment(i int, v interface{}) {
	if p.err != nil {
		return
	}

	p.err = Segment(p.path, i, v)
}

// Sequent operates the same as the package-level function Sequent.
func (p *Parth) Sequent(key string, v interface{}) {
	p.SubSeg(key, 0, v)
}

// Span operates the same as the package-level function Span.
func (p *Parth) Span(i, j int) string {
	if p.err != nil {
		return ""
	}

	s, err := Span(p.path, i, j)
	p.err = err

	return s
}

// SubSeg operates the same as the package-level function SubSeg.
func (p *Parth) SubSeg(key string, i int, v interface{}) {
	if p.err != nil {
		return
	}

	p.err = SubSeg(p.path, key, i, v)
}

// SubSpan operates the same as the package-level function SubSpan.
func (p *Parth) SubSpan(key string, i, j int) string {
	if p.err != nil {
		return ""
	}

	s, err := SubSpan(p.path, key, i, j)
	p.err = err

	return s
}

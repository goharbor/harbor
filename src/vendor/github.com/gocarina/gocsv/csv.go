// Copyright 2014 Jonathan Picques. All rights reserved.
// Use of this source code is governed by a MIT license
// The license can be found in the LICENSE file.

// The GoCSV package aims to provide easy CSV serialization and deserialization to the golang programming language

package gocsv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
)

// FailIfUnmatchedStructTags indicates whether it is considered an error when there is an unmatched
// struct tag.
var FailIfUnmatchedStructTags = false

// FailIfDoubleHeaderNames indicates whether it is considered an error when a header name is repeated
// in the csv header.
var FailIfDoubleHeaderNames = false

// ShouldAlignDuplicateHeadersWithStructFieldOrder indicates whether we should align duplicate CSV
// headers per their alignment in the struct definition.
var ShouldAlignDuplicateHeadersWithStructFieldOrder = false

// TagName defines key in the struct field's tag to scan
var TagName = "csv"

// TagSeparator defines seperator string for multiple csv tags in struct fields
var TagSeparator = ","

// Normalizer is a function that takes and returns a string. It is applied to
// struct and header field values before they are compared. It can be used to alter
// names for comparison. For instance, you could allow case insensitive matching
// or convert '-' to '_'.
type Normalizer func(string) string

type ErrorHandler func(*csv.ParseError) bool

// normalizeName function initially set to a nop Normalizer.
var normalizeName = DefaultNameNormalizer()

// DefaultNameNormalizer is a nop Normalizer.
func DefaultNameNormalizer() Normalizer { return func(s string) string { return s } }

// SetHeaderNormalizer sets the normalizer used to normalize struct and header field names.
func SetHeaderNormalizer(f Normalizer) {
	normalizeName = f
	// Need to clear the cache hen the header normalizer changes.
	structInfoCache = sync.Map{}
}

// --------------------------------------------------------------------------
// CSVWriter used to format CSV

var selfCSVWriter = DefaultCSVWriter

// DefaultCSVWriter is the default SafeCSVWriter used to format CSV (cf. csv.NewWriter)
func DefaultCSVWriter(out io.Writer) *SafeCSVWriter {
	writer := NewSafeCSVWriter(csv.NewWriter(out))

	// As only one rune can be defined as a CSV separator, we are going to trim
	// the custom tag separator and use the first rune.
	if runes := []rune(strings.TrimSpace(TagSeparator)); len(runes) > 0 {
		writer.Comma = runes[0]
	}

	return writer
}

// SetCSVWriter sets the SafeCSVWriter used to format CSV.
func SetCSVWriter(csvWriter func(io.Writer) *SafeCSVWriter) {
	selfCSVWriter = csvWriter
}

func getCSVWriter(out io.Writer) *SafeCSVWriter {
	return selfCSVWriter(out)
}

// --------------------------------------------------------------------------
// CSVReader used to parse CSV

var selfCSVReader = DefaultCSVReader

// DefaultCSVReader is the default CSV reader used to parse CSV (cf. csv.NewReader)
func DefaultCSVReader(in io.Reader) CSVReader {
	return csv.NewReader(in)
}

// LazyCSVReader returns a lazy CSV reader, with LazyQuotes and TrimLeadingSpace.
func LazyCSVReader(in io.Reader) CSVReader {
	csvReader := csv.NewReader(in)
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	return csvReader
}

// SetCSVReader sets the CSV reader used to parse CSV.
func SetCSVReader(csvReader func(io.Reader) CSVReader) {
	selfCSVReader = csvReader
}

func getCSVReader(in io.Reader) CSVReader {
	return selfCSVReader(in)
}

// --------------------------------------------------------------------------
// Marshal functions

// MarshalFile saves the interface as CSV in the file.
func MarshalFile(in interface{}, file *os.File) (err error) {
	return Marshal(in, file)
}

// MarshalString returns the CSV string from the interface.
func MarshalString(in interface{}) (out string, err error) {
	bufferString := bytes.NewBufferString(out)
	if err := Marshal(in, bufferString); err != nil {
		return "", err
	}
	return bufferString.String(), nil
}

// MarshalBytes returns the CSV bytes from the interface.
func MarshalBytes(in interface{}) (out []byte, err error) {
	bufferString := bytes.NewBuffer(out)
	if err := Marshal(in, bufferString); err != nil {
		return nil, err
	}
	return bufferString.Bytes(), nil
}

// Marshal returns the CSV in writer from the interface.
func Marshal(in interface{}, out io.Writer) (err error) {
	writer := getCSVWriter(out)
	return writeTo(writer, in, false)
}

// MarshalWithoutHeaders returns the CSV in writer from the interface.
func MarshalWithoutHeaders(in interface{}, out io.Writer) (err error) {
	writer := getCSVWriter(out)
	return writeTo(writer, in, true)
}

// MarshalChan returns the CSV read from the channel.
func MarshalChan(c <-chan interface{}, out CSVWriter) error {
	return writeFromChan(out, c)
}

// MarshalCSV returns the CSV in writer from the interface.
func MarshalCSV(in interface{}, out CSVWriter) (err error) {
	return writeTo(out, in, false)
}

// MarshalCSVWithoutHeaders returns the CSV in writer from the interface.
func MarshalCSVWithoutHeaders(in interface{}, out CSVWriter) (err error) {
	return writeTo(out, in, true)
}

// --------------------------------------------------------------------------
// Unmarshal functions

// UnmarshalFile parses the CSV from the file in the interface.
func UnmarshalFile(in *os.File, out interface{}) error {
	return Unmarshal(in, out)
}

// UnmarshalFile parses the CSV from the file in the interface.
func UnmarshalFileWithErrorHandler(in *os.File, errHandler ErrorHandler, out interface{}) error {
	return UnmarshalWithErrorHandler(in, errHandler, out)
}

// UnmarshalString parses the CSV from the string in the interface.
func UnmarshalString(in string, out interface{}) error {
	return Unmarshal(strings.NewReader(in), out)
}

// UnmarshalBytes parses the CSV from the bytes in the interface.
func UnmarshalBytes(in []byte, out interface{}) error {
	return Unmarshal(bytes.NewReader(in), out)
}

// Unmarshal parses the CSV from the reader in the interface.
func Unmarshal(in io.Reader, out interface{}) error {
	return readTo(newSimpleDecoderFromReader(in), out)
}

// Unmarshal parses the CSV from the reader in the interface.
func UnmarshalWithErrorHandler(in io.Reader, errHandle ErrorHandler, out interface{}) error {
	return readToWithErrorHandler(newSimpleDecoderFromReader(in), errHandle, out)
}

// UnmarshalWithoutHeaders parses the CSV from the reader in the interface.
func UnmarshalWithoutHeaders(in io.Reader, out interface{}) error {
	return readToWithoutHeaders(newSimpleDecoderFromReader(in), out)
}

// UnmarshalCSVWithoutHeaders parses a headerless CSV with passed in CSV reader
func UnmarshalCSVWithoutHeaders(in CSVReader, out interface{}) error {
	return readToWithoutHeaders(csvDecoder{in}, out)
}

// UnmarshalDecoder parses the CSV from the decoder in the interface
func UnmarshalDecoder(in Decoder, out interface{}) error {
	return readTo(in, out)
}

// UnmarshalCSV parses the CSV from the reader in the interface.
func UnmarshalCSV(in CSVReader, out interface{}) error {
	return readTo(csvDecoder{in}, out)
}

// UnmarshalCSVToMap parses a CSV of 2 columns into a map.
func UnmarshalCSVToMap(in CSVReader, out interface{}) error {
	decoder := NewSimpleDecoderFromCSVReader(in)
	header, err := decoder.getCSVRow()
	if err != nil {
		return err
	}
	if len(header) != 2 {
		return fmt.Errorf("maps can only be created for csv of two columns")
	}
	outValue, outType := getConcreteReflectValueAndType(out)
	if outType.Kind() != reflect.Map {
		return fmt.Errorf("cannot use " + outType.String() + ", only map supported")
	}
	keyType := outType.Key()
	valueType := outType.Elem()
	outValue.Set(reflect.MakeMap(outType))
	for {
		key := reflect.New(keyType)
		value := reflect.New(valueType)
		line, err := decoder.getCSVRow()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if err := setField(key, line[0], false); err != nil {
			return err
		}
		if err := setField(value, line[1], false); err != nil {
			return err
		}
		outValue.SetMapIndex(key.Elem(), value.Elem())
	}
	return nil
}

// UnmarshalToChan parses the CSV from the reader and send each value in the chan c.
// The channel must have a concrete type.
func UnmarshalToChan(in io.Reader, c interface{}) error {
	if c == nil {
		return fmt.Errorf("goscv: channel is %v", c)
	}
	return readEach(newSimpleDecoderFromReader(in), c)
}

// UnmarshalToChanWithoutHeaders parses the CSV from the reader and send each value in the chan c.
// The channel must have a concrete type.
func UnmarshalToChanWithoutHeaders(in io.Reader, c interface{}) error {
	if c == nil {
		return fmt.Errorf("goscv: channel is %v", c)
	}
	return readEachWithoutHeaders(newSimpleDecoderFromReader(in), c)
}

// UnmarshalDecoderToChan parses the CSV from the decoder and send each value in the chan c.
// The channel must have a concrete type.
func UnmarshalDecoderToChan(in SimpleDecoder, c interface{}) error {
	if c == nil {
		return fmt.Errorf("goscv: channel is %v", c)
	}
	return readEach(in, c)
}

// UnmarshalStringToChan parses the CSV from the string and send each value in the chan c.
// The channel must have a concrete type.
func UnmarshalStringToChan(in string, c interface{}) error {
	return UnmarshalToChan(strings.NewReader(in), c)
}

// UnmarshalBytesToChan parses the CSV from the bytes and send each value in the chan c.
// The channel must have a concrete type.
func UnmarshalBytesToChan(in []byte, c interface{}) error {
	return UnmarshalToChan(bytes.NewReader(in), c)
}

// UnmarshalToCallback parses the CSV from the reader and send each value to the given func f.
// The func must look like func(Struct).
func UnmarshalToCallback(in io.Reader, f interface{}) error {
	valueFunc := reflect.ValueOf(f)
	t := reflect.TypeOf(f)
	if t.NumIn() != 1 {
		return fmt.Errorf("the given function must have exactly one parameter")
	}
	cerr := make(chan error)
	c := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, t.In(0)), 0)
	go func() {
		cerr <- UnmarshalToChan(in, c.Interface())
	}()
	for {
		select {
		case err := <-cerr:
			return err
		default:
		}
		v, notClosed := c.Recv()
		if !notClosed || v.Interface() == nil {
			break
		}
		valueFunc.Call([]reflect.Value{v})
	}
	return nil
}

// UnmarshalDecoderToCallback parses the CSV from the decoder and send each value to the given func f.
// The func must look like func(Struct).
func UnmarshalDecoderToCallback(in SimpleDecoder, f interface{}) error {
	valueFunc := reflect.ValueOf(f)
	t := reflect.TypeOf(f)
	if t.NumIn() != 1 {
		return fmt.Errorf("the given function must have exactly one parameter")
	}
	cerr := make(chan error)
	c := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, t.In(0)), 0)
	go func() {
		cerr <- UnmarshalDecoderToChan(in, c.Interface())
	}()
	for {
		select {
		case err := <-cerr:
			return err
		default:
		}
		v, notClosed := c.Recv()
		if !notClosed || v.Interface() == nil {
			break
		}
		valueFunc.Call([]reflect.Value{v})
	}
	return nil
}

// UnmarshalBytesToCallback parses the CSV from the bytes and send each value to the given func f.
// The func must look like func(Struct).
func UnmarshalBytesToCallback(in []byte, f interface{}) error {
	return UnmarshalToCallback(bytes.NewReader(in), f)
}

// UnmarshalStringToCallback parses the CSV from the string and send each value to the given func f.
// The func must look like func(Struct).
func UnmarshalStringToCallback(in string, c interface{}) (err error) {
	return UnmarshalToCallback(strings.NewReader(in), c)
}

// UnmarshalToCallbackWithError parses the CSV from the reader and
// send each value to the given func f.
//
// If func returns error, it will stop processing, drain the
// parser and propagate the error to caller.
//
// The func must look like func(Struct) error.
func UnmarshalToCallbackWithError(in io.Reader, f interface{}) error {
	valueFunc := reflect.ValueOf(f)
	t := reflect.TypeOf(f)
	if t.NumIn() != 1 {
		return fmt.Errorf("the given function must have exactly one parameter")
	}
	if t.NumOut() != 1 {
		return fmt.Errorf("the given function must have exactly one return value")
	}
	if !isErrorType(t.Out(0)) {
		return fmt.Errorf("the given function must only return error.")
	}

	cerr := make(chan error)
	c := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, t.In(0)), 0)
	go func() {
		cerr <- UnmarshalToChan(in, c.Interface())
	}()

	var fErr error
	for {
		select {
		case err := <-cerr:
			if err != nil {
				return err
			}
			return fErr
		default:
		}
		v, notClosed := c.Recv()
		if !notClosed || v.Interface() == nil {
			if err := <- cerr; err != nil {
				fErr = err
			}
			break
		}

		// callback f has already returned an error, stop processing but keep draining the chan c
		if fErr != nil {
			continue
		}

		results := valueFunc.Call([]reflect.Value{v})

		// If the callback f returns an error, stores it and returns it in future.
		errValue := results[0].Interface()
		if errValue != nil {
			fErr = errValue.(error)
		}
	}
	return fErr
}

// UnmarshalBytesToCallbackWithError parses the CSV from the bytes and
// send each value to the given func f.
//
// If func returns error, it will stop processing, drain the
// parser and propagate the error to caller.
//
// The func must look like func(Struct) error.
func UnmarshalBytesToCallbackWithError(in []byte, f interface{}) error {
	return UnmarshalToCallbackWithError(bytes.NewReader(in), f)
}

// UnmarshalStringToCallbackWithError parses the CSV from the string and
// send each value to the given func f.
//
// If func returns error, it will stop processing, drain the
// parser and propagate the error to caller.
//
// The func must look like func(Struct) error.
func UnmarshalStringToCallbackWithError(in string, c interface{}) (err error) {
	return UnmarshalToCallbackWithError(strings.NewReader(in), c)
}

// CSVToMap creates a simple map from a CSV of 2 columns.
func CSVToMap(in io.Reader) (map[string]string, error) {
	decoder := newSimpleDecoderFromReader(in)
	header, err := decoder.getCSVRow()
	if err != nil {
		return nil, err
	}
	if len(header) != 2 {
		return nil, fmt.Errorf("maps can only be created for csv of two columns")
	}
	m := make(map[string]string)
	for {
		line, err := decoder.getCSVRow()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		m[line[0]] = line[1]
	}
	return m, nil
}

// CSVToMaps takes a reader and returns an array of dictionaries, using the header row as the keys
func CSVToMaps(reader io.Reader) ([]map[string]string, error) {
	r := csv.NewReader(reader)
	rows := []map[string]string{}
	var header []string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if header == nil {
			header = record
		} else {
			dict := map[string]string{}
			for i := range header {
				dict[header[i]] = record[i]
			}
			rows = append(rows, dict)
		}
	}
	return rows, nil
}

// CSVToChanMaps parses the CSV from the reader and send a dictionary in the chan c, using the header row as the keys.
func CSVToChanMaps(reader io.Reader, c chan<- map[string]string) error {
	r := csv.NewReader(reader)
	var header []string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header == nil {
			header = record
		} else {
			dict := map[string]string{}
			for i := range header {
				dict[header[i]] = record[i]
			}
			c <- dict
		}
	}
	return nil
}

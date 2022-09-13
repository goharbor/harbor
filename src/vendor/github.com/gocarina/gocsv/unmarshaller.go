package gocsv

import (
	"encoding/csv"
	"fmt"
	"reflect"
)

// Unmarshaller is a CSV to struct unmarshaller.
type Unmarshaller struct {
	reader                 *csv.Reader
	Headers                []string
	fieldInfoMap           []*fieldInfo
	MismatchedHeaders      []string
	MismatchedStructFields []string
	outType                reflect.Type
}

// NewUnmarshaller creates an unmarshaller from a csv.Reader and a struct.
func NewUnmarshaller(reader *csv.Reader, out interface{}) (*Unmarshaller, error) {
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}
	headers = normalizeHeaders(headers)

	um := &Unmarshaller{reader: reader, outType: reflect.TypeOf(out)}
	err = validate(um, out, headers)
	if err != nil {
		return nil, err
	}
	return um, nil
}

// Read returns an interface{} whose runtime type is the same as the struct that
// was used to create the Unmarshaller.
func (um *Unmarshaller) Read() (interface{}, error) {
	row, err := um.reader.Read()
	if err != nil {
		return nil, err
	}
	return um.unmarshalRow(row, nil)
}

// ReadUnmatched is same as Read(), but returns a map of the columns that didn't match a field in the struct
func (um *Unmarshaller) ReadUnmatched() (interface{}, map[string]string, error) {
	row, err := um.reader.Read()
	if err != nil {
		return nil, nil, err
	}
	unmatched := make(map[string]string)
	value, err := um.unmarshalRow(row, unmatched)
	return value, unmatched, err
}

// validate ensures that a struct was used to create the Unmarshaller, and validates
// CSV headers against the CSV tags in the struct.
func validate(um *Unmarshaller, s interface{}, headers []string) error {
	concreteType := reflect.TypeOf(s)
	if concreteType.Kind() == reflect.Ptr {
		concreteType = concreteType.Elem()
	}
	if err := ensureOutInnerType(concreteType); err != nil {
		return err
	}
	structInfo := getStructInfo(concreteType) // Get struct info to get CSV annotations.
	if len(structInfo.Fields) == 0 {
		return ErrNoStructTags
	}
	csvHeadersLabels := make([]*fieldInfo, len(headers)) // Used to store the corresponding header <-> position in CSV
	headerCount := map[string]int{}
	for i, csvColumnHeader := range headers {
		curHeaderCount := headerCount[csvColumnHeader]
		if fieldInfo := getCSVFieldPosition(csvColumnHeader, structInfo, curHeaderCount); fieldInfo != nil {
			csvHeadersLabels[i] = fieldInfo
			if ShouldAlignDuplicateHeadersWithStructFieldOrder {
				curHeaderCount++
				headerCount[csvColumnHeader] = curHeaderCount
			}
		}
	}

	if FailIfDoubleHeaderNames {
		if err := maybeDoubleHeaderNames(headers); err != nil {
			return err
		}
	}

	um.Headers = headers
	um.fieldInfoMap = csvHeadersLabels
	um.MismatchedHeaders = mismatchHeaderFields(structInfo.Fields, headers)
	um.MismatchedStructFields = mismatchStructFields(structInfo.Fields, headers)
	return nil
}

// unmarshalRow converts a CSV row to a struct, based on CSV struct tags.
// If unmatched is non nil, it is populated with any columns that don't map to a struct field
func (um *Unmarshaller) unmarshalRow(row []string, unmatched map[string]string) (interface{}, error) {
	isPointer := false
	concreteOutType := um.outType
	if um.outType.Kind() == reflect.Ptr {
		isPointer = true
		concreteOutType = concreteOutType.Elem()
	}
	outValue := createNewOutInner(isPointer, concreteOutType)
	for j, csvColumnContent := range row {
		if j < len(um.fieldInfoMap) && um.fieldInfoMap[j] != nil {
			fieldInfo := um.fieldInfoMap[j]
			if err := setInnerField(&outValue, isPointer, fieldInfo.IndexChain, csvColumnContent, fieldInfo.omitEmpty); err != nil { // Set field of struct
				return nil, fmt.Errorf("cannot assign field at %v to %s through index chain %v: %v", j, outValue.Type(), fieldInfo.IndexChain, err)
			}
		} else if unmatched != nil {
			unmatched[um.Headers[j]] = csvColumnContent
		}
	}
	return outValue.Interface(), nil
}

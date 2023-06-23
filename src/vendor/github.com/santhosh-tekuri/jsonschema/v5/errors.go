package jsonschema

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"
)

// InvalidJSONTypeError is the error type returned by ValidateInterface.
// this tells that specified go object is not valid jsonType.
type InvalidJSONTypeError string

func (e InvalidJSONTypeError) Error() string {
	return fmt.Sprintf("jsonschema: invalid jsonType: %s", string(e))
}

// InfiniteLoopError is returned by Compile/Validate.
// this gives url#keywordLocation that lead to infinity loop.
type InfiniteLoopError string

func (e InfiniteLoopError) Error() string {
	return "jsonschema: infinite loop " + string(e)
}

func infiniteLoopError(stack []schemaRef, sref schemaRef) InfiniteLoopError {
	var path string
	for _, ref := range stack {
		if path == "" {
			path += ref.schema.Location
		} else {
			path += "/" + ref.path
		}
	}
	return InfiniteLoopError(path + "/" + sref.path)
}

// SchemaError is the error type returned by Compile.
type SchemaError struct {
	// SchemaURL is the url to json-schema that filed to compile.
	// This is helpful, if your schema refers to external schemas
	SchemaURL string

	// Err is the error that occurred during compilation.
	// It could be ValidationError, because compilation validates
	// given schema against the json meta-schema
	Err error
}

func (se *SchemaError) Unwrap() error {
	return se.Err
}

func (se *SchemaError) Error() string {
	s := fmt.Sprintf("jsonschema %s compilation failed", se.SchemaURL)
	if se.Err != nil {
		return fmt.Sprintf("%s: %v", s, strings.TrimPrefix(se.Err.Error(), "jsonschema: "))
	}
	return s
}

func (se *SchemaError) GoString() string {
	if _, ok := se.Err.(*ValidationError); ok {
		return fmt.Sprintf("jsonschema %s compilation failed\n%#v", se.SchemaURL, se.Err)
	}
	return se.Error()
}

// ValidationError is the error type returned by Validate.
type ValidationError struct {
	KeywordLocation         string             // validation path of validating keyword or schema
	AbsoluteKeywordLocation string             // absolute location of validating keyword or schema
	InstanceLocation        string             // location of the json value within the instance being validated
	Message                 string             // describes error
	Causes                  []*ValidationError // nested validation errors
}

func (ve *ValidationError) add(causes ...error) error {
	for _, cause := range causes {
		ve.Causes = append(ve.Causes, cause.(*ValidationError))
	}
	return ve
}

func (ve *ValidationError) causes(err error) error {
	if err := err.(*ValidationError); err.Message == "" {
		ve.Causes = err.Causes
	} else {
		ve.add(err)
	}
	return ve
}

func (ve *ValidationError) leaf() *ValidationError {
	if strings.HasSuffix(ve.KeywordLocation, "/anyOf") || strings.HasSuffix(ve.KeywordLocation, "/oneOf") {
		if len(ve.Causes) == 1 {
			return ve.Causes[0].leaf()
		}
		return ve
	}
	if len(ve.Causes) > 0 {
		return ve.Causes[0].leaf()
	}
	return ve
}

func (ve *ValidationError) Error() string {
	err := ve.leaf()
	u, _ := split(ve.AbsoluteKeywordLocation)
	return fmt.Sprintf("jsonschema: %s does not validate with %s: %s", quote(err.InstanceLocation), u+"#"+err.KeywordLocation, err.Message)
}

func (ve *ValidationError) GoString() string {
	sloc := ve.AbsoluteKeywordLocation
	sloc = sloc[strings.IndexByte(sloc, '#')+1:]
	msg := fmt.Sprintf("[I#%s] [S#%s] %s", ve.InstanceLocation, sloc, ve.Message)
	for _, c := range ve.Causes {
		for _, line := range strings.Split(c.GoString(), "\n") {
			msg += "\n  " + line
		}
	}
	return msg
}

func joinPtr(ptr1, ptr2 string) string {
	if len(ptr1) == 0 {
		return ptr2
	}
	if len(ptr2) == 0 {
		return ptr1
	}
	return ptr1 + "/" + ptr2
}

func quote(s string) string {
	var w = bytes.NewBuffer(make([]byte, 0, len(s)+10))
	w.WriteByte('\'')
	start := 0
	for i := 0; i < len(s); {
		b := s[i]
		if b < 0x20 {
			if start < i {
				w.WriteString(s[start:i])
			}
			switch b {
			case '\n':
				w.WriteString(`\n`)
			case '\r':
				w.WriteString(`\r`)
			case '\f':
				w.WriteString(`\f`)
			case '\t':
				w.WriteString(`\t`)
			default:
				w.WriteString(`\u00`)
				w.Write(hex(b >> 4))
				w.Write(hex(b & 0xF))
			}
			i++
			start = i
			continue
		}
		if b < utf8.RuneSelf {
			if b == '\\' || b == '"' {
				if start < i {
					w.WriteString(s[start:i])
				}
				switch b {
				case '\\':
					w.WriteString(`\\`)
				case '\'':
					w.WriteString(`\'`)
				}
				i++
				start = i
				continue
			}
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			if start < i {
				w.WriteString(s[start:i])
			}
			w.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		if r == '\u2028' || r == '\u2029' {
			if start < i {
				w.WriteString(s[start:i])
			}
			w.WriteString(`\u202`)
			w.Write(hex(uint8(r & 0xF)))
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		w.WriteString(s[start:])
	}
	w.WriteByte('\'')
	return w.String()
}

var hexBytes = []byte("0123456789abcdef")

func hex(i uint8) []byte {
	return hexBytes[i : i+1]
}

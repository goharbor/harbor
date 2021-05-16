// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

// mysqlInterpolate parses query and replace all "?" with encoded args.
// If there are more "?" than len(args), returns ErrMissingArgs.
// Otherwise, if there are less "?" than len(args), the redundant args are omitted.
func mysqlInterpolate(query string, args ...interface{}) (string, error) {
	return mysqlLikeInterpolate(MySQL, query, args...)
}

func mysqlLikeInterpolate(flavor Flavor, query string, args ...interface{}) (string, error) {
	// Roughly estimate the size to avoid useless memory allocation and copy.
	buf := make([]byte, 0, len(query)+len(args)*20)

	var quote rune
	var err error
	cnt := 0
	max := len(args)
	escaping := false
	offset := 0
	target := query
	r, sz := utf8.DecodeRuneInString(target)

	for ; sz != 0; r, sz = utf8.DecodeRuneInString(target) {
		offset += sz
		target = query[offset:]

		if escaping {
			escaping = false
			continue
		}

		switch r {
		case '?':
			if quote != 0 {
				continue
			}

			if cnt >= max {
				return "", ErrInterpolateMissingArgs
			}

			buf = append(buf, query[:offset-sz]...)
			buf, err = encodeValue(buf, args[cnt], flavor)

			if err != nil {
				return "", err
			}

			query = target
			offset = 0
			cnt++

		case '\'':
			if quote == '\'' {
				quote = 0
				continue
			}

			if quote == 0 {
				quote = '\''
			}

		case '"':
			if quote == '"' {
				quote = 0
				continue
			}

			if quote == 0 {
				quote = '"'
			}

		case '`':
			if quote == '`' {
				quote = 0
				continue
			}

			if quote == 0 {
				quote = '`'
			}

		case '\\':
			if quote != 0 {
				escaping = true
			}
		}
	}

	buf = append(buf, query...)
	return *(*string)(unsafe.Pointer(&buf)), nil
}

// postgresqlInterpolate parses query and replace all "$*" with encoded args.
// If there are more "$*" than len(args), returns ErrMissingArgs.
// Otherwise, if there are less "$*" than len(args), the redundant args are omitted.
func postgresqlInterpolate(query string, args ...interface{}) (string, error) {
	// Roughly estimate the size to avoid useless memory allocation and copy.
	buf := make([]byte, 0, len(query)+len(args)*20)

	var quote rune
	var dollarQuote string
	var err error
	var idx int64
	max := len(args)
	escaping := false
	offset := 0
	target := query
	r, sz := utf8.DecodeRuneInString(target)

	for ; sz != 0; r, sz = utf8.DecodeRuneInString(target) {
		offset += sz
		target = query[offset:]

		if escaping {
			escaping = false
			continue
		}

		switch r {
		case '$':
			if quote != 0 {
				if quote != '$' {
					continue
				}

				// Try to find the end of dollar quote.
				pos := offset

				for r, sz = utf8.DecodeRuneInString(target); sz != 0 && r != '$'; r, sz = utf8.DecodeRuneInString(target) {
					pos += sz
					target = query[pos:]
				}

				if sz == 0 {
					break
				}

				if r == '$' {
					dq := query[offset : pos+sz]
					offset = pos
					target = query[offset:]

					if dq == dollarQuote {
						quote = 0
						dollarQuote = ""
						offset += sz
						target = query[offset:]
					}

					continue
				}

				continue
			}

			oldSz := sz
			pos := offset
			r, sz = utf8.DecodeRuneInString(target)

			if '1' <= r && r <= '9' {
				// A placeholder is found.
				pos += sz
				target = query[pos:]

				for r, sz = utf8.DecodeRuneInString(target); sz != 0 && '0' <= r && r <= '9'; r, sz = utf8.DecodeRuneInString(target) {
					pos += sz
					target = query[pos:]
				}

				idx, err = strconv.ParseInt(query[offset:pos], 10, strconv.IntSize)

				if err != nil {
					return "", err
				}

				if int(idx) >= max+1 {
					return "", ErrInterpolateMissingArgs
				}

				buf = append(buf, query[:offset-oldSz]...)
				buf, err = encodeValue(buf, args[idx-1], PostgreSQL)

				if err != nil {
					return "", err
				}

				query = target
				offset = 0

				if sz == 0 {
					break
				}

				continue
			}

			// Try to find the beginning of dollar quote.
			for ; sz != 0 && r != '$' && unicode.IsLetter(r); r, sz = utf8.DecodeRuneInString(target) {
				pos += sz
				target = query[pos:]
			}

			if sz == 0 {
				break
			}

			if !unicode.IsLetter(r) && r != '$' {
				continue
			}

			pos += sz
			quote = '$'
			dollarQuote = query[offset:pos]
			offset = pos
			target = query[offset:]

		case '\'':
			if quote == '\'' {
				// PostgreSQL uses two single quotes to represent one single quote.
				r, sz = utf8.DecodeRuneInString(target)

				if r == '\'' {
					offset += sz
					target = query[offset:]
					continue
				}

				quote = 0
				continue
			}

			if quote == 0 {
				quote = '\''
			}

		case '"':
			if quote == '"' {
				quote = 0
				continue
			}

			if quote == 0 {
				quote = '"'
			}

		case '\\':
			if quote == '\'' || quote == '"' {
				escaping = true
			}
		}
	}

	buf = append(buf, query...)
	return *(*string)(unsafe.Pointer(&buf)), nil
}

// mysqlInterpolate works the same as MySQL interpolating.
func sqliteInterpolate(query string, args ...interface{}) (string, error) {
	return mysqlLikeInterpolate(SQLite, query, args...)
}

func encodeValue(buf []byte, arg interface{}, flavor Flavor) ([]byte, error) {
	switch v := arg.(type) {
	case nil:
		buf = append(buf, "NULL"...)

	case bool:
		if v {
			buf = append(buf, "TRUE"...)
		} else {
			buf = append(buf, "FALSE"...)
		}

	case int:
		buf = strconv.AppendInt(buf, int64(v), 10)

	case int8:
		buf = strconv.AppendInt(buf, int64(v), 10)

	case int16:
		buf = strconv.AppendInt(buf, int64(v), 10)

	case int32:
		buf = strconv.AppendInt(buf, int64(v), 10)

	case int64:
		buf = strconv.AppendInt(buf, v, 10)

	case uint:
		buf = strconv.AppendUint(buf, uint64(v), 10)

	case uint8:
		buf = strconv.AppendUint(buf, uint64(v), 10)

	case uint16:
		buf = strconv.AppendUint(buf, uint64(v), 10)

	case uint32:
		buf = strconv.AppendUint(buf, uint64(v), 10)

	case uint64:
		buf = strconv.AppendUint(buf, v, 10)

	case float32:
		buf = strconv.AppendFloat(buf, float64(v), 'g', -1, 32)

	case float64:
		buf = strconv.AppendFloat(buf, v, 'g', -1, 64)

	case []byte:
		if v == nil {
			buf = append(buf, "NULL"...)
			break
		}

		switch flavor {
		case MySQL:
			buf = append(buf, "_binary"...)
			buf = quoteStringValue(buf, *(*string)(unsafe.Pointer(&v)), flavor)

		case PostgreSQL:
			buf = append(buf, "E'\\\\x"...)
			buf = appendHex(buf, v)
			buf = append(buf, "'::bytea"...)

		case SQLite:
			buf = append(buf, "X'"...)
			buf = appendHex(buf, v)
			buf = append(buf, '\'')
		}

	case string:
		buf = quoteStringValue(buf, v, flavor)

	case time.Time:
		if v.IsZero() {
			buf = append(buf, "'0000-00-00'"...)
			break
		}

		// In SQL standard, the precision of fractional seconds in time literal is up to 6 digits.
		// Round up v.
		v = v.Add(500 * time.Nanosecond)
		buf = append(buf, '\'')

		switch flavor {
		case MySQL:
			buf = append(buf, v.Format("2006-01-02 15:04:05.999999")...)

		case PostgreSQL:
			buf = append(buf, v.Format("2006-01-02 15:04:05.999999 MST")...)

		case SQLite:
			buf = append(buf, v.Format("2006-01-02 15:04:05.000")...)
		}

		buf = append(buf, '\'')

	case fmt.Stringer:
		buf = quoteStringValue(buf, v.String(), flavor)

	default:
		return nil, ErrInterpolateUnsupportedArgs
	}

	return buf, nil
}

var hexDigits = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'}

func appendHex(buf, v []byte) []byte {
	for _, b := range v {
		buf = append(buf, hexDigits[(b>>4)&0xF], hexDigits[b&0xF])
	}

	return buf
}

func quoteStringValue(buf []byte, s string, flavor Flavor) []byte {
	if flavor == PostgreSQL {
		buf = append(buf, 'E')
	}

	buf = append(buf, '\'')
	r, sz := utf8.DecodeRuneInString(s)

	for ; sz != 0; r, sz = utf8.DecodeRuneInString(s) {
		switch r {
		case '\x00':
			buf = append(buf, "\\0"...)

		case '\b':
			buf = append(buf, "\\b"...)

		case '\n':
			buf = append(buf, "\\n"...)

		case '\r':
			buf = append(buf, "\\r"...)

		case '\t':
			buf = append(buf, "\\t"...)

		case '\x1a':
			buf = append(buf, "\\Z"...)

		case '\'':
			buf = append(buf, "\\'"...)

		case '"':
			buf = append(buf, "\\\""...)

		case '\\':
			buf = append(buf, "\\\\"...)

		default:
			buf = append(buf, s[:sz]...)
		}

		s = s[sz:]
	}

	buf = append(buf, '\'')
	return buf
}

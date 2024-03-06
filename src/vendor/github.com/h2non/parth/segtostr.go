package parth

import (
	"strconv"
	"unicode"
)

func segmentToBool(path string, i int) (bool, error) {
	s, err := segmentToString(path, i)
	if err != nil {
		return false, err
	}

	v, err := strconv.ParseBool(s)
	if err != nil {
		return false, ErrDataUnparsable
	}

	return v, nil
}

func segmentToFloatN(path string, i, size int) (float64, error) {
	ss, err := segmentToString(path, i)
	if err != nil {
		return 0.0, err
	}

	s, ok := firstFloatFromString(ss)
	if !ok {
		return 0.0, err
	}

	v, err := strconv.ParseFloat(s, size)
	if err != nil {
		return 0.0, ErrDataUnparsable
	}

	return v, nil
}

func segmentToIntN(path string, i, size int) (int64, error) {
	ss, err := segmentToString(path, i)
	if err != nil {
		return 0, err
	}

	s, ok := firstIntFromString(ss)
	if !ok {
		return 0, ErrDataUnparsable
	}

	v, err := strconv.ParseInt(s, 10, size)
	if err != nil {
		return 0, ErrDataUnparsable
	}

	return v, nil
}

func segmentToString(path string, i int) (string, error) {
	j := i + 1
	if i < 0 {
		i--
	}

	s, err := Span(path, i, j)
	if err != nil {
		return "", err
	}

	if s[0] == '/' {
		s = s[1:]
	}

	return s, nil
}

func segmentToUintN(path string, i, size int) (uint64, error) {
	ss, err := segmentToString(path, i)
	if err != nil {
		return 0, err
	}

	s, ok := firstUintFromString(ss)
	if !ok {
		return 0, ErrDataUnparsable
	}

	v, err := strconv.ParseUint(s, 10, size)
	if err != nil {
		return 0, ErrDataUnparsable
	}

	return v, nil
}

func subSegToBool(path, key string, i int) (bool, error) {
	s, err := subSegToString(path, key, i)
	if err != nil {
		return false, err
	}

	v, err := strconv.ParseBool(s)
	if err != nil {
		return false, ErrDataUnparsable
	}

	return v, nil
}

func subSegToFloatN(path, key string, i, size int) (float64, error) {
	ss, err := subSegToString(path, key, i)
	if err != nil {
		return 0.0, err
	}

	s, ok := firstFloatFromString(ss)
	if !ok {
		return 0.0, ErrDataUnparsable
	}

	v, err := strconv.ParseFloat(s, size)
	if err != nil {
		return 0.0, ErrDataUnparsable
	}

	return v, nil
}

func subSegToIntN(path, key string, i, size int) (int64, error) {
	ss, err := subSegToString(path, key, i)
	if err != nil {
		return 0, err
	}

	s, ok := firstIntFromString(ss)
	if !ok {
		return 0, ErrDataUnparsable
	}

	v, err := strconv.ParseInt(s, 10, size)
	if err != nil {
		return 0, ErrDataUnparsable
	}

	return v, nil
}

func subSegToString(path, key string, i int) (string, error) {
	ki, ok := segIndexByKey(path, key)
	if !ok {
		return "", ErrKeySegNotFound
	}

	i++

	s, err := segmentToString(path[ki:], i)
	if err != nil {
		return "", err
	}

	return s, nil
}

func subSegToUintN(path, key string, i, size int) (uint64, error) {
	ss, err := subSegToString(path, key, i)
	if err != nil {
		return 0, err
	}

	s, ok := firstUintFromString(ss)
	if !ok {
		return 0, ErrDataUnparsable
	}

	v, err := strconv.ParseUint(s, 10, size)
	if err != nil {
		return 0, ErrDataUnparsable
	}

	return v, nil
}

func firstUintFromString(s string) (string, bool) {
	ind, l := 0, 0

	for n := 0; n < len(s); n++ {
		if unicode.IsDigit(rune(s[n])) {
			if l == 0 {
				ind = n
			}

			l++
		} else {
			if l == 0 && s[n] == '.' {
				if n+1 < len(s) && unicode.IsDigit(rune(s[n+1])) {
					return "0", true
				}

				break
			}

			if l > 0 {
				break
			}
		}
	}

	if l == 0 {
		return "", false
	}

	return s[ind : ind+l], true
}

func firstIntFromString(s string) (string, bool) { //nolint
	ind, l := 0, 0

	for n := 0; n < len(s); n++ {
		if unicode.IsDigit(rune(s[n])) {
			if l == 0 {
				ind = n
			}

			l++
		} else if s[n] == '-' {
			if l == 0 {
				ind = n
				l++
			} else {
				break
			}
		} else {
			if l == 0 && s[n] == '.' {
				if n+1 < len(s) && unicode.IsDigit(rune(s[n+1])) {
					return "0", true
				}

				break
			}

			if l > 0 {
				break
			}
		}
	}

	if l == 0 {
		return "", false
	}

	return s[ind : ind+l], true
}

func firstFloatFromString(s string) (string, bool) { //nolint
	c, ind, l := 0, 0, 0

	for n := 0; n < len(s); n++ {
		if unicode.IsDigit(rune(s[n])) {
			if l == 0 {
				ind = n
			}

			l++
		} else if s[n] == '-' {
			if l == 0 {
				ind = n
				l++
			} else {
				break
			}
		} else if s[n] == '.' {
			if l == 0 {
				ind = n
			}

			if c > 0 {
				break
			}

			l++
			c++
		} else if s[n] == 'e' && l > 0 && n+1 < len(s) && s[n+1] == '+' {
			l++
		} else if s[n] == '+' && l > 0 && s[n-1] == 'e' {
			if n+1 < len(s) && unicode.IsDigit(rune(s[n+1])) {
				l++
				continue
			}

			l--
			break
		} else {
			if l > 0 {
				break
			}
		}
	}

	if l == 0 || s[ind:ind+l] == "." {
		return "", false
	}

	return s[ind : ind+l], true
}

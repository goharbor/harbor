// Package strutil provides various utilities for manipulating strings
package strutil

import (
	"bytes"
	"regexp"

	"github.com/mattn/go-runewidth"
)

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

// PadRight returns a new string of a specified length in which the end of the current string is padded with spaces or with a specified Unicode character.
func PadRight(str string, length int, pad byte) string {
	slen := StringWidth(str)
	if slen >= length {
		return str
	}
	buf := bytes.NewBufferString(str)
	for i := 0; i < length-slen; i++ {
		buf.WriteByte(pad)
	}
	return buf.String()
}

// PadLeft returns a new string of a specified length in which the beginning of the current string is padded with spaces or with a specified Unicode character.
func PadLeft(str string, length int, pad byte) string {
	slen := StringWidth(str)
	if slen >= length {
		return str
	}
	var buf bytes.Buffer
	for i := 0; i < length-slen; i++ {
		buf.WriteByte(pad)
	}
	buf.WriteString(str)
	return buf.String()
}

// Resize resizes the string with the given length. It ellipses with '...' when the string's length exceeds
// the desired length or pads spaces to the right of the string when length is smaller than desired
func Resize(s string, length uint, rightAlign bool) string {
	slen := StringWidth(s)
	n := int(length)
	if slen == n {
		return s
	}
	// Pads only when length of the string smaller than len needed
	if rightAlign {
		s = PadLeft(s, n, ' ')
	} else {
		s = PadRight(s, n, ' ')
	}
	if slen > n {
		rs := []rune(s)
		var buf bytes.Buffer
		w := 0
		for _, r := range rs {
			buf.WriteRune(r)
			rw := RuneWidth(r)
			if w+rw >= n-3 {
				break
			}
			w += rw
		}
		buf.WriteString("...")
		s = buf.String()
	}
	return s
}

// Join joins the list of the string with the delim provided.
// Returns an empty string for empty list
func Join(list []string, delim string) string {
	if len(list) == 0 {
		return ""
	}
	var buf bytes.Buffer
	for i := 0; i < len(list)-1; i++ {
		buf.WriteString(list[i] + delim)
	}
	buf.WriteString(list[len(list)-1])
	return buf.String()
}

// Strip strips the string of all colors
func Strip(s string) string {
	return re.ReplaceAllString(s, "")
}

// StringWidth returns the actual width of the string without colors
func StringWidth(s string) int {
	return runewidth.StringWidth(Strip(s))
}

// RuneWidth returns the actual width of the rune
func RuneWidth(s rune) int {
	return runewidth.RuneWidth(s)
}

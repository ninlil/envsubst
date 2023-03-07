package envsubst

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

var (
	prefix rune = '$'
	start  rune = '('
	end    rune = ')'
)

type errMissing string

func (e errMissing) Error() string {
	return fmt.Sprintf("field '%s' is missing", string(e))
}

// SetPrefix can change the default '$' character
// Valid characters are '$', '%', '#' and '&'
func SetPrefix(ch rune) bool {
	switch ch {
	case '$', '%', '&', '#':
		prefix = ch
	default:
		return false
	}
	return true
}

// SetWrapper can change the default '{}'
// Valid options are (any of) '()', '{}', '[]', '<>'
func SetWrapper(ch rune) bool {
	switch ch {
	case '(', ')':
		start = '('
		end = ')'
	case '{', '}':
		start = '{'
		end = '}'
	case '[', ']':
		start = '['
		end = ']'
	case '<', '>':
		start = '<'
		end = '>'
	default:
		return false
	}
	return true
}

// LookupEnv get from os.Env (missing variables fails the process)
func LookupEnv(name string) (string, bool) {
	return os.LookupEnv(name)
}

// Getenv get from os.Env (missing variables becomes empty)
func Getenv(name string) (string, bool) {
	return os.Getenv(name), true
}

// Map converts a map[string]string into a mapping-function for the Convert*-functions
func Map(fields map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		val, ok := fields[name]
		return val, ok
	}
}

// ConvertString takes a string and converts variables using a mpaaing-function (like Map, Getenv or LookupEnv)
func ConvertString(str string, mapping func(string) (string, bool)) (string, error) {
	if len(str) == 0 {
		return str, nil
	}

	out, err := ConvertBytes([]byte(str), mapping)

	return string(out), err
}

// ConvertBytes takes a byte-array and converts variables using a mpaaing-function (like Map, Getenv or LookupEnv)
func ConvertBytes(data []byte, mapping func(string) (string, bool)) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	buf := bytes.NewReader(data)
	out := bytes.Buffer{}

	err := Convert(buf, &out, mapping)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

// Convert does a steam-conversion of variables using a mpaaing-function (like Map, Getenv or LookupEnv)
func Convert(rd io.Reader, wr io.Writer, mapping func(string) (string, bool)) error {
	var ch rune
	var err error
	var state int
	var varname string

	if mapping == nil {
		mapping = Getenv
	}

	bufrd := bufio.NewReader(rd)
	bufwr := bufio.NewWriter(wr)
	defer bufwr.Flush()

	for {
		ch, _, err = bufrd.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch true {
		case ch == prefix && state == 0:
			state++

		case ch == start && state == 1:
			state++
			varname = ""

		case ch == end && state == 2:
			if varname != "" {
				match, found := mapping(varname)
				if !found {
					return errMissing(varname)
				}
				_, _ = bufwr.WriteString(match)
			} else {
				_, _ = bufwr.WriteRune(prefix)
				_, _ = bufwr.WriteRune(start)
				_, _ = bufwr.WriteRune(end)
			}
			varname = ""
			state = 0

		default:
			switch state {
			case 2:
				varname += string(ch)

			case 1:
				_, _ = bufwr.WriteRune(prefix)
				if ch != prefix {
					_, _ = bufwr.WriteRune(ch)
					state = 0
				}

			default:
				_, _ = bufwr.WriteRune(ch)
			}
		}
	}
	switch state {
	case 2:
		_, _ = bufwr.WriteRune(prefix)
		_, _ = bufwr.WriteRune(start)
		_, _ = bufwr.WriteString(varname)

	case 1:
		_, _ = bufwr.WriteRune(prefix)
	}

	return nil
}

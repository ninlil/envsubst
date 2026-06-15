package envsubst

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

const (
	defaultPrefix rune = '$'
	defaultStart  rune = '('
	defaultEnd    rune = ')'
)

// Replacer is a struct that can be used to get a concurrent version of the Convert*-functions (see Replacer.ConvertString, Replacer.ConvertBytes and Replacer.Convert)
type Replacer struct {
	prefix rune
	start  rune
	end    rune
}

var defaultReplacer = NewReplacer()

type errMissing string

func (e errMissing) Error() string {
	return fmt.Sprintf("field '%s' is missing", string(e))
}

// NewReplacer creates a new Replacer with the default settings (prefix '$' and wrapper '{}') and applies the given options (like WithPrefix and WithWrapper)
func NewReplacer(options ...func(*Replacer)) *Replacer {
	r := &Replacer{
		prefix: defaultPrefix,
		start:  defaultStart,
		end:    defaultEnd,
	}
	for _, option := range options {
		option(r)
	}
	return r
}

// WithPrefix is an option for NewReplacer to set the prefix character
func WithPrefix(ch rune) func(*Replacer) {
	return func(r *Replacer) {
		r.SetPrefix(ch)
	}
}

// WithWrapper is an option for NewReplacer to set the wrapper characters
func WithWrapper(ch rune) func(*Replacer) {
	return func(r *Replacer) {
		r.SetWrapper(ch)
	}
}

// SetPrefix can change the default '$' character
// Valid characters are '$', '%', '#' and '&'
func SetPrefix(ch rune) bool {
	return defaultReplacer.SetPrefix(ch)
}

// SetPrefix can change the default '$' character
// Valid characters are '$', '%', '#' and '&'
func (r *Replacer) SetPrefix(ch rune) bool {
	switch ch {
	case '$', '%', '&', '#':
		r.prefix = ch
	default:
		return false
	}
	return true
}

// SetWrapper can change the default '{}'
// Valid options are (any of) '()', '{}', '[]', '<>'
func SetWrapper(ch rune) bool {
	return defaultReplacer.SetWrapper(ch)
}

// SetWrapper can change the default '{}'
// Valid options are (any of) '()', '{}', '[]', '<>'
func (r *Replacer) SetWrapper(ch rune) bool {
	switch ch {
	case '(', ')':
		r.start = '('
		r.end = ')'
	case '{', '}':
		r.start = '{'
		r.end = '}'
	case '[', ']':
		r.start = '['
		r.end = ']'
	case '<', '>':
		r.start = '<'
		r.end = '>'
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

// ConvertString takes a string and converts variables using a mapping-function (like Map, Getenv or LookupEnv)
func ConvertString(str string, mapping func(string) (string, bool)) (string, error) {
	return defaultReplacer.ConvertString(str, mapping)
}

// ConvertString takes a string and converts variables using a mapping-function (like Map, Getenv or LookupEnv)
func (r *Replacer) ConvertString(str string, mapping func(string) (string, bool)) (string, error) {
	if len(str) == 0 {
		return str, nil
	}

	out, err := r.ConvertBytes([]byte(str), mapping)

	return string(out), err
}

// ConvertBytes takes a byte-array and converts variables using a mapping-function (like Map, Getenv or LookupEnv)
func ConvertBytes(data []byte, mapping func(string) (string, bool)) ([]byte, error) {
	return defaultReplacer.ConvertBytes(data, mapping)
}

// ConvertBytes takes a byte-array and converts variables using a mapping-function (like Map, Getenv or LookupEnv)
func (r *Replacer) ConvertBytes(data []byte, mapping func(string) (string, bool)) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	buf := bytes.NewReader(data)
	out := bytes.Buffer{}

	err := r.Convert(buf, &out, mapping)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

// Convert does a stream-conversion of variables using a mapping-function (like Map, Getenv or LookupEnv)
func Convert(rd io.Reader, wr io.Writer, mapping func(string) (string, bool)) error {
	return defaultReplacer.Convert(rd, wr, mapping)
}

// Convert does a stream-conversion of variables using a mapping-function (like Map, Getenv or LookupEnv)
func (r *Replacer) Convert(rd io.Reader, wr io.Writer, mapping func(string) (string, bool)) (err error) {
	var ch rune
	var state int
	var varname string

	if mapping == nil {
		mapping = Getenv
	}

	bufrd := bufio.NewReader(rd)
	bufwr := bufio.NewWriter(wr)
	defer func() {
		if ferr := bufwr.Flush(); ferr != nil && err == nil {
			err = ferr
		}
	}()

	for {
		ch, _, err = bufrd.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch true {
		case ch == r.prefix && state == 0:
			state++

		case ch == r.start && state == 1:
			state++
			varname = ""

		case ch == r.end && state == 2:
			if varname != "" {
				match, found := mapping(varname)
				if !found {
					return errMissing(varname)
				}
				_, _ = bufwr.WriteString(match)
			} else {
				_, _ = bufwr.WriteRune(r.prefix)
				_, _ = bufwr.WriteRune(r.start)
				_, _ = bufwr.WriteRune(r.end)
			}
			varname = ""
			state = 0

		default:
			switch state {
			case 2:
				varname += string(ch)

			case 1:
				_, _ = bufwr.WriteRune(r.prefix)
				if ch != r.prefix {
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
		_, _ = bufwr.WriteRune(r.prefix)
		_, _ = bufwr.WriteRune(r.start)
		_, _ = bufwr.WriteString(varname)

	case 1:
		_, _ = bufwr.WriteRune(r.prefix)
	}

	return nil
}

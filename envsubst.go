package envsubst

import (
	"bufio"
	"bytes"
	"io"
)

func ConvertString(str string, fields map[string]string) (string, error) {
	if len(fields) == 0 || len(str) == 0 {
		return str, nil
	}

	out, err := ConvertBytes([]byte(str), fields)

	return string(out), err
}

func ConvertBytes(data []byte, fields map[string]string) ([]byte, error) {
	if len(fields) == 0 || len(data) == 0 {
		return data, nil
	}

	buf := bytes.NewReader(data)
	out := bytes.Buffer{}

	err := Convert(buf, &out, fields)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func Convert(rd io.Reader, wr *bytes.Buffer, fields map[string]string) error {
	var ch rune
	var err error
	var state int
	var varname string

	brd := bufio.NewReader(rd)
	for {
		ch, _, err = brd.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch true {
		case ch == '$' && state == 0:
			state++
		case ch == '(' && state == 1:
			state++
			varname = ""
		case ch == ')' && state == 2:
			if varname != "" {
				wr.WriteString(fields[varname])
			} else {
				wr.WriteString("$()")
			}
			varname = ""
			state = 0
		default:
			switch state {
			case 2:
				varname += string(ch)
			case 1:
				wr.WriteRune('$')
				if ch != '$' {
					wr.WriteRune(ch)
					state = 0
				}
			default:
				wr.WriteRune(ch)
			}
		}
	}
	switch state {
	case 2:
		wr.WriteString("$(")
		wr.WriteString(varname)
	case 1:
		wr.WriteRune('$')
	}

	return nil
}

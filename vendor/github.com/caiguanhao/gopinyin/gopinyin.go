// Small utils for Chinese Pinyin.
package gopinyin

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
)

type Pinyins []string

// Get the first letter of all pinyins.
func (pys Pinyins) Abbreviate() (abbreviated Pinyins) {
	for _, py := range pys {
		abbreviated = append(abbreviated, py[0:1])
	}
	return
}

// Find out possible matches for abbreviated pinyin.
func (pys Pinyins) Expand() (out Pinyins) {
	for _, i := range pys {
		out = append(out, _MAP[i])
	}
	return
}

// Convert the pinyin array to a string.
func (pys Pinyins) Join(bytes ...byte) (out string) {
	out = strings.Join(pys, string(bytes))
	return
}

// Returns a compiled regular expression.
func (pys Pinyins) Regexp() (out *regexp.Regexp) {
	out = regexp.MustCompile(pys.RegexpString())
	return
}

// Returns a regular expression string.
func (pys Pinyins) RegexpString() (out string) {
	for _, py := range pys {
		out += `\^` + py + "[a-z]*"
	}
	return
}

// Convert the expanded pinyins to WHERE SQL statement for PostgreSQL.
func (pys Pinyins) SQL(column string) (out string) {
	var ret []string
	for _, py := range pys {
		ret = append(ret, fmt.Sprintf("'{%s}'", py))
	}
	if len(ret) == 0 {
		return
	}
	out = fmt.Sprintf("SEQUENCED_ARRAY_CONTAINS(%s, %s)", column, strings.Join(ret, ", "))
	return
}

// Convert to SQL driver value.
func (pys Pinyins) Value() (value driver.Value, err error) {
	var buf []byte
	var ret = []byte{'{'}
	for _, py := range pys {
		for _, b := range []byte(py) {
			if b >= 65 && b <= 90 { // A - Z
				b += 32
			}
			if b >= 97 && b <= 122 { // a - z
				buf = append(buf, b)
			}
		}
		if len(buf) > 0 {
			buf = append(buf, ',')
			ret = append(ret, buf...)
			buf = nil
		}
	}
	if ret[len(ret)-1] == ',' {
		ret[len(ret)-1] = '}'
	} else {
		ret = append(ret, '}')
	}
	value = driver.Value(string(ret))
	return
}

// Split a pinyin string into an array.
func Split(in string) (out Pinyins) {
	_in := []byte(in)
	var buf []byte
	for i, b := range _in {
		if b >= 65 && b <= 90 { // A - Z
			b += 32
		}
		if b >= 97 && b <= 122 { // a - z
			buf = append(buf, b)
		}
		if i < len(_in)-1 && _MAP[string(append(buf, _in[i+1]))] != "" {
			continue
		}
		if buf != nil {
			if _MAP[string(buf)] != "" {
				out = append(out, string(buf))
			}
			buf = nil
		}
	}
	return
}

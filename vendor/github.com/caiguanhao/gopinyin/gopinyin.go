// Small utils for Chinese Pinyin.
package gopinyin

import (
	"fmt"
	"strings"
)

type Pinyins []string

// Find out possible matches for abbreviated pinyin.
func (pys Pinyins) Expand() (out Pinyins) {
	for _, i := range pys {
		out = append(out, _MAP[i])
	}
	return
}

// Convert the expanded pinyins to WHERE SQL statement for PostgreSQL.
func (pys Pinyins) SQL(column string) (out string) {
	var ret []string
	for _, py := range pys {
		ret = append(ret, fmt.Sprintf("%s && '{%s}'", column, py))
	}
	out = strings.Join(ret, " AND ")
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

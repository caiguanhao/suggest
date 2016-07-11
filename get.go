package suggest

import (
	"errors"
	"fmt"

	"github.com/caiguanhao/gopinyin"
)

func query(pys gopinyin.Pinyins, abbr string) (stmt string, args []interface{}) {
	stmt = "SELECT id, word, pinyin, sogou_count FROM suggestions"

	if len(pys) == 1 {
		if len(pys[0]) == 1 {
			stmt += " WHERE abbr ~~ $1 ORDER BY length ASC, sogou_count DESC LIMIT 10"
			args = []interface{}{"%" + pys[0] + "%"}
			return
		}

		stmt += " WHERE pinyin ~~ $1 ORDER BY length ASC, sogou_count DESC LIMIT 10"
		args = []interface{}{"%^" + pys[0] + "%"}
		return
	}

	stmt += " WHERE abbr ~~ $1 AND pinyin ~ $2 ORDER BY SCORE(abbr, $3) DESC, sogou_count DESC LIMIT 10"
	args = []interface{}{"%" + abbr + "%", pys.RegexpString(), abbr}
	return
}

func (suggest Suggest) get(pinyin string) (rets []map[string]*interface{}, pys gopinyin.Pinyins, err error) {
	pys = gopinyin.Split(pinyin)
	abbr := pys.Abbreviate().Join()
	if abbr == "" {
		err = errors.New("please enter valid pinyins or pinyin abbreviations")
		return
	}

	stmt, args := query(pys, abbr)
	rets, err = suggest.Query(stmt, args...)
	return
}

func (suggest Suggest) serializeGet(rets []map[string]*interface{}, pys gopinyin.Pinyins, _err error) (suggestions []map[string]interface{}, err error) {
	if err = _err; err != nil {
		return
	}
	pinyinRegexp := pys.Regexp()
	for _, item := range rets {
		start, end, pinyin := -1, -1, (*item["pinyin"]).([]byte)
		if pos := pinyinRegexp.FindIndex(pinyin); pos != nil {
			start = 0
			for i := 0; i < pos[0]; i++ {
				if pinyin[i] == '^' {
					start++
				}
			}
			end = start + len(pys)
		}
		suggestions = append(suggestions, map[string]interface{}{
			"start": start,
			"end":   end,
			"text":  fmt.Sprintf("%s", *item["word"]),
		})
	}
	return
}

func (suggest Suggest) Get(pinyin string) (err error) {
	var rets []map[string]*interface{}

	rets, _, err = suggest.get(pinyin)
	if err != nil {
		return
	}

	for _, item := range rets {
		fmt.Printf("%s\n", *item["word"])
	}
	return
}

package suggest

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/caiguanhao/gopinyin"
	"github.com/caiguanhao/gotogether"
	"github.com/caiguanhao/searchresults"
)

func (suggest Suggest) get(pinyin string) (rets []map[string]*interface{}, err error) {
	pys := gopinyin.Split(pinyin)
	abbr := pys.Abbreviate().Join()
	if abbr == "" {
		err = errors.New("please enter valid pinyins or pinyin abbreviations")
		return
	}
	rets, err = suggest.Query(
		"SELECT id, word, sogou_count FROM data WHERE abbr ~~ $1 AND CONTAINS(pinyin, $2) ORDER BY SCORE(abbr, $3) DESC, sogou_count DESC LIMIT 10",
		fmt.Sprintf("%%%s%%", abbr), pys, abbr,
	)
	return
}

func (suggest Suggest) Get(pinyin string) (err error) {
	var rets []map[string]*interface{}

	rets, err = suggest.get(pinyin)
	if err != nil {
		return
	}

	var noCount []interface{}
	for _, ret := range rets {
		if (*ret["sogou_count"]).(int64) > 0 {
			continue
		}
		noCount = append(noCount, ret)
	}

	if len(noCount) > 0 {
		err = suggest.BulkExec("UPDATE data SET sogou_count = $1 WHERE ID = $2", func(stmt *sql.Stmt) (err error) {
			gotogether.Enumerable(noCount).Queue(func(item interface{}) {
				data := item.(map[string]*interface{})
				id := *data["id"]
				word := fmt.Sprintf("%s", *data["word"])
				count, _ := searchresults.GetSogouCount(word)
				_, err = stmt.Exec(count, id)
			}).WithConcurrency(5).Run()
			return
		})
		if err != nil {
			return
		}

		rets, err = suggest.get(pinyin)
		if err != nil {
			return
		}

	}

	for _, item := range rets {
		fmt.Printf("%s - %d\n", *item["word"], *item["sogou_count"])
	}
	return
}

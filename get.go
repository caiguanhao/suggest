package suggest

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/caiguanhao/gopinyin"
	"github.com/caiguanhao/gotogether"
	"github.com/caiguanhao/searchresults"
)

func (suggest Suggest) get(pinyin string) (rets []interface{}, err error) {
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
	var rets []interface{}

	rets, err = suggest.get(pinyin)
	if err != nil {
		return
	}

	noCount := gotogether.Enumerable(rets).Filter(func(item interface{}) bool {
		data := item.([]interface{})
		count := (*(data[2].(*interface{}))).(int64)
		if count > 0 {
			return false
		}
		return true
	})

	if len(noCount) > 0 {
		err = suggest.BulkExec("UPDATE data SET sogou_count = $1 WHERE ID = $2", func(stmt *sql.Stmt) (err error) {
			noCount.Queue(func(item interface{}) {
				data := item.([]interface{})
				id := (*(data[0].(*interface{}))).(int64)
				word := fmt.Sprintf("%s", *(data[1].(*interface{})))
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
		data := item.([]interface{})
		word := *(data[1].(*interface{}))
		count := (*(data[2].(*interface{}))).(int64)
		fmt.Printf("%s - %d\n", word, count)
	}
	return
}

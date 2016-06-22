package suggest

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/caiguanhao/sogoudict"
)

// Download dictionary file and save its content into database.
func (suggest Suggest) GetDict(ID int) (err error) {
	var req *http.Request
	url := fmt.Sprintf("http://download.pinyin.sogou.com/dict/download_cell.php?id=%d&name=", ID)
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Referer", "http://pinyin.sogou.com/")
	var resp *http.Response
	resp, err = (&http.Client{}).Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	fmt.Fprintf(os.Stderr, "dict %d downloaded (%d bytes)\n", ID, len(body))

	var dict sogoudict.SogouDict
	dict, err = sogoudict.Parse(bytes.NewReader(body))

	fmt.Fprintf(os.Stderr, "found %d items in dict %d\n", len(dict.Items), ID)

	err = suggest.BulkInsert(func(db *sql.DB) bool {
		var count int
		db.QueryRow("SELECT count(*) FROM data WHERE sogou_id = $1", ID).Scan(&count)
		if count > 0 {
			fmt.Println(ID, "already added")
			return false
		}
		return true
	}, func(stmt *sql.Stmt) (err error) {
		for _, item := range dict.Items {
			_, err = stmt.Exec(StringArray(item.Pinyin), item.Text, ID, 0)
			if err != nil {
				return
			}
		}
		return
	}, "data", "pinyin", "word", "sogou_id", "weight")
	return
}

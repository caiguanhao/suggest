package suggest

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/caiguanhao/sogoudict"
)

func join(in []string) (out string) {
	out = "^" + strings.Join(in, "^")
	return
}

// Download dictionary file and save its content into database.
func (suggest Suggest) GetDict(ID int) (err error) {
	var req *http.Request
	url := fmt.Sprintf("http://download.pinyin.sogou.com/dict/download_cell.php?id=%d&name=", ID)

	fmt.Fprintf(os.Stderr, "downloading dict %d from %s\n", ID, url)

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
			_, err = stmt.Exec(join(item.Pinyin), strings.Join(item.Abbr, ""), item.Text, ID)
			if err != nil {
				return
			}
		}
		return
	}, "data", "pinyin", "abbr", "word", "sogou_id")

	if err == nil {
		fmt.Fprintf(os.Stderr, "%d items added to database\n", len(dict.Items))
	}

	return
}

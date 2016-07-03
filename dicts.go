package suggest

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/caiguanhao/sogoudict"
)

type reader struct {
	reader   io.Reader
	read     int64
	last     time.Time
	progress func(read int64)
}

func (r *reader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.read += int64(n)
	now := time.Now()
	timediff := float64(now.Sub(r.last).Nanoseconds()) / 1e9
	if timediff > 0.1 {
		r.progress(r.read)
		r.last = now
	}
	return
}

func join(in []string) (out string) {
	out = "^" + strings.Join(in, "^")
	return
}

// Download dictionary file and save its content into database.
func (suggest Suggest) GetDict(ID int, out func(period string, done, total int64, format string, a ...interface{})) (err error) {
	var req *http.Request
	url := fmt.Sprintf("http://download.pinyin.sogou.com/dict/download_cell.php?id=%d&name=", ID)

	out("downloading", 0, 0, "downloading dict %d from %s\n", ID, url)

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
	total := resp.ContentLength
	defer resp.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(&reader{
		reader: resp.Body,
		progress: func(read int64) {
			out("downloading", read, total, "")
		},
	})
	if err != nil {
		return
	}

	out("downloaded", total, total, "dict %d downloaded (%d bytes)\n", ID, total)

	var dict sogoudict.SogouDict
	dict, err = sogoudict.Parse(bytes.NewReader(body))
	if err != nil {
		return
	}

	total = int64(len(dict.Items))

	out("importing", total, total, "found %d items in dict %d\n", total, ID)

	err = suggest.BulkInsert(func(db *sql.DB) bool {
		var count int64
		db.QueryRow("SELECT count(*) FROM suggestions WHERE sogou_id = $1", ID).Scan(&count)
		if count > 0 {
			out("imported", count, count, "%d already added\n", ID)
			return false
		}
		return true
	}, func(stmt *sql.Stmt) (err error) {
		last := time.Now()
		for i, item := range dict.Items {
			_, err = stmt.Exec(join(item.Pinyin), strings.Join(item.Abbr, ""), item.Text, ID)
			if err != nil {
				return
			}
			now := time.Now()
			timediff := float64(now.Sub(last).Nanoseconds()) / 1e9
			if timediff > 0.5 {
				out("importing", int64(i+1), total, "")
				last = now
			}
		}
		return
	}, "suggestions", "pinyin", "abbr", "word", "sogou_id")

	if err == nil {
		out("imported", total, total, "%d items added to database\n", total)
	} else {
		return
	}

	err = suggest.BulkExec(
		"UPDATE dicts SET suggestion_count = (SELECT COUNT(*) FROM suggestions WHERE sogou_id = $1) WHERE sogou_id = $1",
		func(stmt *sql.Stmt) (err error) {
			_, err = stmt.Exec(ID)
			return
		},
	)

	return
}

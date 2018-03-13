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
	read     int
	last     time.Time
	progress func(read int)
}

func (r *reader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.read += n
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
func (suggest Suggest) GetDict(ID int, out func(id int, format string, a ...interface{}), progress func(id, done, total int)) (total int, err error) {
	if out == nil {
		out = func(id int, format string, a ...interface{}) {}
	}
	if progress == nil {
		progress = func(id, done, total int) {}
	}

	var req *http.Request
	url := fmt.Sprintf("https://pinyin.sogou.com/d/dict/download_cell.php?id=%d&name=dict", ID)

	out(ID, "[downloading] downloading dict %d from %s\n", ID, url)
	progress(ID, 0, 0)

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
	total = int(resp.ContentLength)
	defer resp.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(&reader{
		reader: resp.Body,
		progress: func(read int) {
			progress(ID, read, total)
		},
	})
	if err != nil {
		return
	}

	out(ID, "[downloaded] dict %d downloaded (%d bytes)\n", ID, total)
	progress(ID, total, total)

	var dict sogoudict.SogouDict
	dict, err = sogoudict.Parse(bytes.NewReader(body))
	if err != nil {
		return
	}

	total = len(dict.Items)

	out(ID, "[importing] found %d items in dict %d\n", total, ID)

	err = suggest.BulkExec("INSERT INTO suggestions (pinyin, abbr, word, length, sogou_id) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (word) DO NOTHING",
		func(stmt *sql.Stmt) (err error) {
			last := time.Now()
			for i, item := range dict.Items {
				_, err = stmt.Exec(join(item.Pinyin), strings.Join(item.Abbr, ""), item.Text, len(item.Pinyin), ID)
				if err != nil {
					return
				}
				now := time.Now()
				timediff := float64(now.Sub(last).Nanoseconds()) / 1e9
				if timediff > 0.1 {
					progress(ID, i+1, total)
					last = now
				}
			}
			return
		})

	if err == nil {
		out(ID, "[imported] %d items added to database\n", total)
		progress(ID, total, total)
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

	if err != nil {
		return
	}

	var c map[string]*interface{}
	c, err = suggest.QueryOne("SELECT suggestion_count FROM dicts WHERE sogou_id = $1", ID)
	if err == nil && c != nil {
		total = int((*c["suggestion_count"]).(int64))
	}

	return
}

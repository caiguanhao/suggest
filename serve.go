package suggest

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/caiguanhao/suggest/web"
	"github.com/urfave/cli"
)

var useLocalHtml bool

func (suggest Suggest) Serve(c *cli.Context) (err error) {
	useLocalHtml = c.Bool("local")

	http.HandleFunc("/suggestions", func(resp http.ResponseWriter, req *http.Request) {
		query := req.URL.Query().Get("q")
		rets, err := suggest.get(query)
		if err != nil {
			printErr(resp, err)
			return
		}
		var suggestions []string
		for _, item := range rets {
			suggestions = append(suggestions, fmt.Sprintf("%s", *item["word"]))
		}
		printJson(resp, suggestions, suggestions == nil)
	})

	http.HandleFunc("/lists", func(resp http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.Header.Get("Accept"), "application/json") {
			c, err := suggest.Query("SELECT count(*) FROM dicts")
			if err != nil {
				printErr(resp, err)
				return
			}
			resp.Header().Set("Total-Items", fmt.Sprintf("%d", *c[0]["count"]))
			per, _, offset := paginate(req)
			dicts, err := suggest.Query(
				"SELECT dicts.id, dicts.name, dicts.download_count, dicts.sogou_id, dicts.updated_at, categories.name as category_name FROM dicts "+
					"LEFT JOIN categories ON categories.id = dicts.category_id "+
					"ORDER BY download_count DESC LIMIT $1 OFFSET $2", per, offset)
			if err != nil {
				printErr(resp, err)
				return
			}
			printJson(resp, dicts, dicts == nil)
			return
		}

		serveHtml(resp, req, "web/lists.html")
	})

	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		serveHtml(resp, req, "web/index.html")
	})

	err = http.ListenAndServe(":8080", nil)
	return
}

func paginate(req *http.Request) (per, page, offset int) {
	per, _ = strconv.Atoi(req.URL.Query().Get("per"))
	if per < 1 || per > 100 {
		per = 10
	}
	page, _ = strconv.Atoi(req.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	offset = per * (page - 1)
	return
}

func printJson(resp http.ResponseWriter, content interface{}, isNil bool) {
	resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	if isNil {
		resp.Write([]byte{'[', ']'})
		return
	}
	if err := json.NewEncoder(resp).Encode(content); err != nil {
		printErr(resp, err)
	}
}

func printErr(resp http.ResponseWriter, err error) {
	resp.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(resp).Encode(map[string]string{
		"error": err.Error(),
	})
}

func serveHtml(resp http.ResponseWriter, req *http.Request, filename string) {
	if useLocalHtml {
		http.ServeFile(resp, req, filename)
		return
	}

	resp.Header().Set("Content-Type", "text/html; charset=utf-8")

	if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		resp.Header().Set("Content-Encoding", "gzip")
		writer := gzip.NewWriter(resp)
		writer.Write(web.Files[filename])
		writer.Flush()
		return
	}

	resp.Write(web.Files[filename])
}

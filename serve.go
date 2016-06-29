package suggest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (suggest Suggest) Serve() (err error) {
	http.HandleFunc("/suggestions", func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/json")
		query := req.URL.Query().Get("q")
		rets, err := suggest.get(query)
		if err != nil {
			resp.Write([]byte{'[', ']'})
			return
		}
		var suggestions []string
		for _, item := range rets {
			suggestions = append(suggestions, fmt.Sprintf("%s", *item["word"]))
		}
		if suggestions == nil {
			resp.Write([]byte{'[', ']'})
			return
		}
		json.NewEncoder(resp).Encode(suggestions)
	})

	http.HandleFunc("/lists", func(resp http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.Header.Get("Accept"), "application/json") {
			c, err := suggest.Query("SELECT count(*) FROM dicts")
			if err == nil {
				resp.Header().Set("Total-Items", fmt.Sprintf("%d", *c[0]["count"]))
			}
			per, _ := strconv.Atoi(req.URL.Query().Get("per"))
			if per < 1 || per > 100 {
				per = 10
			}
			page, _ := strconv.Atoi(req.URL.Query().Get("page"))
			if page < 1 {
				page = 1
			}
			offset := per * (page - 1)
			rets, err := suggest.Query(
				"SELECT dicts.id, dicts.name, dicts.download_count, dicts.sogou_id, dicts.updated_at, categories.name as category_name FROM dicts "+
					"LEFT JOIN categories ON categories.id = dicts.category_id "+
					"ORDER BY download_count DESC LIMIT $1 OFFSET $2", per, offset)
			if err != nil || rets == nil {
				resp.Write([]byte{'[', ']'})
				return
			}
			json.NewEncoder(resp).Encode(rets)
			return
		}

		serve(resp, req, "web/lists.html")
	})

	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		serve(resp, req, "web/index.html")
	})

	err = http.ListenAndServe(":8080", nil)
	return
}

func serve(resp http.ResponseWriter, req *http.Request, filename string) {
	// http.ServeFile(resp, req, filename)
	resp.Header().Set("Content-Type", "text/html; charset=utf-8")
	resp.Header().Set("Content-Encoding", "gzip")
	resp.Write(web[filename])
}

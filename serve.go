package suggest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (suggest Suggest) Serve() (err error) {
	http.HandleFunc("/suggestions", func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/json")
		query := req.URL.Query().Get("q")
		rets, err := suggest.get(query)
		if err != nil {
			fmt.Fprintln(resp, "[]")
			return
		}
		var suggestions []string
		for _, item := range rets {
			suggestions = append(suggestions, fmt.Sprintf("%s", *item["word"]))
		}
		json.NewEncoder(resp).Encode(suggestions)
	})
	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		// http.ServeFile(resp, req, "web/index.html")
		resp.Write(web_index_html)
	})
	http.ListenAndServe(":8080", nil)
	return
}

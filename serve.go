package suggest

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/caiguanhao/suggest/web"
	"github.com/gorilla/websocket"
	"github.com/urfave/cli"
)

var (
	upgrader = websocket.Upgrader{
		Error: func(resp http.ResponseWriter, req *http.Request, status int, err error) {
			printErr(resp, err)
		},
	}
	clients        = make(map[*websocket.Conn]bool)
	isGettingLists = false
	isGettingDicts = make(map[int]bool)
	useLocalHtml   = false

	mutex sync.Mutex
)

func (suggest Suggest) Serve(c *cli.Context) (err error) {
	useLocalHtml = c.Bool("local")

	http.HandleFunc("/approximate_suggestion_count", func(resp http.ResponseWriter, req *http.Request) {
		count, _ := suggest.QueryOne("SELECT reltuples AS approximate_suggestion_count FROM pg_class WHERE relname = $1", "suggestions")
		printJson(resp, count, count == nil)
	})

	http.HandleFunc("/suggestions", func(resp http.ResponseWriter, req *http.Request) {
		query := req.URL.Query().Get("q")
		suggestions, err := suggest.serializeGet(suggest.get(query))
		if err != nil {
			printErr(resp, err)
			return
		}
		printJson(resp, suggestions, suggestions == nil)
	})

	http.HandleFunc("/lists", func(resp http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.Header.Get("Accept"), "application/json") {
			count, dicts, err := suggest.CountAndQuery(getListsCountAndQuery(req))
			if err != nil {
				printErr(resp, err)
				return
			}
			resp.Header().Set("Total-Items", fmt.Sprintf("%d", count))
			printJson(resp, dicts, dicts == nil)
			return
		}

		serveHtml(resp, req, "web/lists.html")
	})

	http.HandleFunc("/get", suggest.GetHandler())

	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		serveHtml(resp, req, "web/index.html")
	})

	err = http.ListenAndServe(":8080", nil)
	return
}

func broadcast(content interface{}) {
	if len(clients) == 0 {
		return
	}
	body, err := json.Marshal(content)
	if err != nil {
		return
	}
	for conn := range clients {
		mutex.Lock()
		conn.WriteMessage(websocket.TextMessage, body)
		mutex.Unlock()
	}
}

func getListsProgress(done, total int) {
	broadcast(map[string]interface{}{
		"type":  "get-lists-progress",
		"done":  done,
		"total": total,
	})
}

func getListsBroadcast(format string, a ...interface{}) {
	broadcast(map[string]interface{}{
		"type":        "get-lists",
		"status_text": fmt.Sprintf(format, a...),
	})
}

func getDictsProgress(id, done, total int) {
	broadcast(map[string]interface{}{
		"type":  "get-dicts-progress",
		"value": id,
		"done":  done,
		"total": total,
	})
}

func getDictsDoneBroadcast(id, total int) {
	broadcast(map[string]interface{}{
		"type":  "get-dicts-done",
		"value": id,
		"total": total,
	})
}

func (suggest Suggest) execute(ws *websocket.Conn, msg map[string]string) {
	switch msg["type"] {
	case "get-suggestions":
		suggestions, err := suggest.serializeGet(suggest.get(msg["value"]))
		if err != nil {
			printWSError(ws, err)
			return
		}
		printWSJson(ws, suggestions, suggestions == nil)
	case "get-lists":
		if isGettingLists {
			printWSErrorString(ws, "list getting has already been started, please wait\n")
			return
		}
		isGettingLists = true
		suggest.GetLists(getListsBroadcast, getListsProgress)
		isGettingLists = false
	case "get-dicts":
		id, err := strconv.Atoi(msg["value"])
		if err != nil {
			printWSError(ws, err)
			return
		}
		mutex.Lock()
		getting := isGettingDicts[id]
		mutex.Unlock()
		if getting {
			printWSErrorString(ws, "dict getting has already been started, please wait\n")
			return
		}
		mutex.Lock()
		isGettingDicts[id] = true
		mutex.Unlock()
		if total, err := suggest.GetDict(id, nil, getDictsProgress); err != nil {
			printWSError(ws, err)
			getDictsDoneBroadcast(id, 0)
		} else {
			getDictsDoneBroadcast(id, total)
		}
		mutex.Lock()
		isGettingDicts[id] = false
		mutex.Unlock()
	}
}

func (suggest Suggest) GetHandler() func(resp http.ResponseWriter, req *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		ws, err := upgrader.Upgrade(resp, req, nil)
		if err != nil {
			return
		}
		clients[ws] = true
		fmt.Println("new client:", ws.RemoteAddr().String(), "total clients:", len(clients))
		for {
			var msg map[string]string
			_, reader, err := ws.NextReader()
			if err != nil {
				break
			}
			if err := json.NewDecoder(reader).Decode(&msg); err != nil {
				printWSError(ws, err)
				continue
			}
			go suggest.execute(ws, msg)
		}
		ws.Close()
		if _, ok := clients[ws]; ok {
			delete(clients, ws)
			fmt.Println("deleted client:", ws.RemoteAddr().String(), "total clients:", len(clients))
		}
	}
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
	resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(resp).Encode(map[string]string{
		"error": err.Error(),
	})
}

func printWSJson(ws *websocket.Conn, content interface{}, isNil bool) {
	mutex.Lock()
	defer mutex.Unlock()
	if isNil {
		ws.WriteMessage(websocket.TextMessage, []byte{'[', ']'})
		return
	}
	if err := ws.WriteJSON(content); err != nil {
		printWSError(ws, err)
	}
}

func printWSErrorString(ws *websocket.Conn, err string) {
	mutex.Lock()
	ws.WriteJSON(map[string]string{"error": err})
	mutex.Unlock()
}

func printWSError(ws *websocket.Conn, err error) {
	printWSErrorString(ws, err.Error())
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

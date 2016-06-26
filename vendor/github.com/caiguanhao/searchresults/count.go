// Get total count of search results from various search engines.
package searchresults

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	ErrNoCount = errors.New("failed to get count")
)

var (
	BaiduCountRegexp    = regexp.MustCompile("百度为您找到相关结果约([0-9,]+?)个")
	SogouCountRegexp    = regexp.MustCompile("<!--resultbarnum:([0-9,]+?)-->")
	SoDotComCountRegexp = regexp.MustCompile("找到相关结果约([0-9,]+?)个")

	GetCountRequestUserAgent = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36`
)

func request(url string, re *regexp.Regexp) (count int, err error) {
	var req *http.Request
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", GetCountRequestUserAgent)
	var resp *http.Response
	resp, err = (&http.Client{
		Timeout: time.Second * 3,
	}).Do(req)
	if err != nil {
		return
	}
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	resp.Body.Close()
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		err = ErrNoCount
		return
	}
	count, err = strconv.Atoi(strings.Replace(string(matches[1]), ",", "", -1))
	return
}

// Get total search result count from baidu.com.
func GetBaiduCount(query string) (count int, err error) {
	url := fmt.Sprintf(`https://www.baidu.com/s?wd="%s"`, query)
	count, err = request(url, BaiduCountRegexp)
	return
}

// Get total search result count from sogou.com.
func GetSogouCount(query string) (count int, err error) {
	url := fmt.Sprintf(`https://www.sogou.com/web?ie=utf8&query="%s"`, query)
	count, err = request(url, SogouCountRegexp)
	return
}

// Get total search result count from so.com.
func GetSoDotComCount(query string) (count int, err error) {
	url := fmt.Sprintf(`https://www.so.com/s?q="%s"`, query)
	count, err = request(url, SoDotComCountRegexp)
	return
}

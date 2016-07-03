package suggest

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/caiguanhao/gotogether"
)

// Get dictionary links of all pages of each category.
func (suggest Suggest) GetLists(out func(format string, a ...interface{}), progress func(done, total int)) (err error) {
	if out == nil {
		out = func(format string, a ...interface{}) {}
	}
	if progress == nil {
		progress = func(done, total int) {}
	}

	var doc *goquery.Document

	done, total := 0, 0
	progress(done, total)

	doc, err = goquery.NewDocument("http://pinyin.sogou.com/dict/")
	if err != nil {
		return
	}

	titles := doc.Find(".dict_category_list_title")
	total += titles.Length()
	progress(done, total)
	regexCategoryUrl := regexp.MustCompile("/dict/cate/index/([0-9]+)")
	err = suggest.BulkInsert(nil, func(stmt *sql.Stmt) (err error) {
		titles.Each(func(_ int, s *goquery.Selection) {
			a := s.Find("a")
			name := a.Text()
			href := a.AttrOr("href", "")
			matches := regexCategoryUrl.FindStringSubmatch(href)
			if len(matches) > 1 {
				_, err = stmt.Exec(matches[1], name)
			}
			done += 1
			progress(done, total)
		})
		return
	}, "categories", "sogou_category_id", "name")
	if err != nil {
		return
	}

	out("list of all categories saved\n")

	var rets []map[string]*interface{}
	rets, err = suggest.Query("SELECT id, sogou_category_id, name FROM categories")
	if err != nil {
		return
	}

	var categories []interface{}
	for _, ret := range rets {
		categories = append(categories, ret)
	}

	gotogether.Enumerable(categories).Queue(func(category interface{}) {
		data := category.(map[string]*interface{})
		categoryID := *data["id"]
		sogouCategoryID := *data["sogou_category_id"]
		categoryName := *data["name"]

		doc, err = goquery.NewDocument(fmt.Sprintf("http://pinyin.sogou.com/dict/cate/index/%d/download", sogouCategoryID))
		if err != nil {
			return
		}
		doc.Find("#dict_page_list a").Last().Remove()
		var totalPages int
		totalPages, err = strconv.Atoi(doc.Find("#dict_page_list a").Last().Text())
		if err != nil {
			return
		}

		out("found %d pages of %s\n", totalPages, categoryName)
		total += totalPages
		progress(done, total)

		var urls []interface{}
		for i := 1; i <= totalPages; i++ {
			urls = append(urls, fmt.Sprintf("http://pinyin.sogou.com/dict/cate/index/%d/download/%d", sogouCategoryID, i))
		}

		regexDetailsUrl := regexp.MustCompile("/dict/detail/index/([0-9]+)")
		err = suggest.BulkInsert(nil, func(stmt *sql.Stmt) (err error) {
			gotogether.Enumerable(urls).Queue(func(item interface{}) {
				var doc *goquery.Document
				doc, err = goquery.NewDocument(item.(string))
				if err != nil {
					return
				}
				done += 1
				progress(done, total)
				doc.Find("#dict_detail_list .dict_detail_block").Each(func(_ int, s *goquery.Selection) {
					link := s.Find(".detail_title a")
					href := link.AttrOr("href", "")
					matches := regexDetailsUrl.FindStringSubmatch(href)
					if len(matches) < 2 {
						return
					}
					name := link.Text()
					content := s.Find(".dict_detail_show .show_content")
					examples := content.Eq(0).Text()
					downloadCount := content.Eq(1).Text()
					loc, _ := time.LoadLocation("Asia/Shanghai")
					updatedAt, _ := time.ParseInLocation("2006-01-02 15:04:05", content.Eq(2).Text(), loc)
					_, err = stmt.Exec(matches[1], categoryID, name, downloadCount, examples, updatedAt.UTC())
				})
			}).WithConcurrency(5).Run()
			return
		}, "dicts", "sogou_id", "category_id", "name", "download_count", "examples", "updated_at")

		out("info of all dicts in %s saved\n", categoryName)
	}).WithConcurrency(5).Run()

	out("finished getting lists\n")
	return
}

func (suggest Suggest) GetListsCount() (int64, error) {
	ret, err := suggest.QueryOne("SELECT count(*) FROM dicts")
	if err != nil || ret == nil {
		return 0, err
	}
	return (*ret["count"]).(int64), nil
}

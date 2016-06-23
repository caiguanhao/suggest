package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/caiguanhao/gopinyin"
	"github.com/caiguanhao/suggest"
	"github.com/urfave/cli"
)

func main() {
	s := suggest.Suggest{
		DataSource: "postgres://localhost/suggest?sslmode=disable",
	}
	app := cli.NewApp()
	app.Name = "suggest"
	app.Usage = "Find suggestion."
	app.Commands = []cli.Command{
		{
			Name:  "get-lists",
			Usage: "Get dictionary links of all pages of each category.",
			Action: func(c *cli.Context) error {
				s.GetLists()
				return nil
			},
		},
		{
			Name:  "get-dicts",
			Usage: "Download dictionary file and save its content into database.",
			Action: func(c *cli.Context) error {
				dict, err := strconv.Atoi(c.Args().First())
				if err != nil {
					return err
				}
				s.GetDict(dict)
				return nil
			},
		},
		{
			Name:  "get",
			Usage: "Get suggest for word.",
			Action: func(c *cli.Context) error {
				sql := fmt.Sprintf("SELECT * FROM (SELECT word, %s AS rel FROM data ORDER BY rel DESC LIMIT 20) AS INN WHERE INN.rel > -1",
					gopinyin.Split(c.Args().First()).Expand().SQL("pinyin"))
				rets, err := s.Query(sql)
				if err != nil {
					return err
				}
				for _, ret := range rets {
					retData := ret.([]interface{})
					word := *(retData[0].(*interface{}))
					fmt.Printf("%s\n", word)
				}
				return nil
			},
		},
	}
	app.Run(os.Args)
}

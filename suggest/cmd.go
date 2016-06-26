package main

import (
	"fmt"
	"os"
	"strconv"

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
				return s.GetLists()
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
				return s.GetDict(dict)
			},
		},
		{
			Name:  "get",
			Usage: "Get suggestions for word.",
			Action: func(c *cli.Context) error {
				return s.Get(c.Args().First())
			},
		},
		{
			Name:  "serve",
			Usage: "Run suggestions server.",
			Action: func(c *cli.Context) error {
				return s.Serve()
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

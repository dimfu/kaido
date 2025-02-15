package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dimfu/kaido/store"
	"github.com/urfave/cli/v3"
)

func init() {
	if err := setup(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	store, err := store.GetInstance()
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	cmd := &cli.Command{
		Name:  "kaido",
		Usage: "Collect kaido battle tour time records",
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "collect all or some map records",
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Println("this ran")
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/dimfu/kaido/collectors"
	"github.com/dimfu/kaido/store"
	"github.com/urfave/cli/v3"
)

func init() {
	if err := setup(); err != nil {
		log.Fatalf("failed to initiate setup: %v", err)
	}
	if err := collectors.GenerateLeaderboardTracks(); err != nil {
		if !errors.Is(err, collectors.ERR_ALREADY_GENERATED) {
			log.Fatalf("cannot get leaderboard tracks data: %v", err)
		}
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
					fmt.Println("recoooooooooords")
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

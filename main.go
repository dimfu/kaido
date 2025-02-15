package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dimfu/kaido/collectors"
	"github.com/dimfu/kaido/config"
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
	cfg := config.GetConfig()
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
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "leaderboard",
						Value: "all",
						Usage: "get specific leaderboards eg;kaido run ., default to all",
					},
					// TODO: add optional flag for showing only current month records
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					leaderboardFlag := c.String("leaderboard")
					if len(leaderboardFlag) > 0 {
						// TODO: handle multiple regions input
						lbToLower := strings.ToLower(leaderboardFlag)
						_, exists := cfg.Leaderboards[lbToLower]
						if !exists {
							fmt.Println("cannot find leaderboard")
						}
						// fetch the result of current leaderboard
					}
					return nil
				},
			},
			{
				Name:  "leaderboards",
				Usage: "See all available leaderboards",
				Action: func(ctx context.Context, c *cli.Command) error {
					for region := range cfg.Leaderboards {
						fmt.Println(region)
					}
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

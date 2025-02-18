package main

import (
	"context"
	"log"
	"strings"

	"github.com/dimfu/kaido/collectors"
	"github.com/dimfu/kaido/config"
	"github.com/dimfu/kaido/store"
	"github.com/urfave/cli/v3"
)

func collectLeaderboardRecords(ctx context.Context, c *cli.Command) error {
	s, err := store.GetInstance()
	if err != nil {
		return err
	}

	leaderboardFlag := c.String("leaderboard")
	leaderboard := strings.ToLower(leaderboardFlag)

	// TODO: handle all region at once
	if leaderboard == "all" {
		// get all leaderboard record from all region
		return nil
	}

	timing := collectors.TimingTable{
		Store: s,
		Cfg:   config.GetConfig(),
	}

	// TODO: handle multiple regions input
	results, err := timing.Extract(leaderboard)
	if err != nil {
		return err
	}

	// TODO: compare previous and current top time
	for key, result := range results {
		log.Println(key, result.Prev, result.Curr)
	}

	return nil
}

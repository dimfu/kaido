package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

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
	cfg := config.GetConfig()
	leaderboardFlag := c.String("leaderboard")
	if len(leaderboardFlag) > 0 {
		// TODO: handle multiple regions input
		lbToLower := strings.ToLower(leaderboardFlag)
		leaderboard, exists := cfg.Leaderboards[lbToLower]
		if !exists {
			fmt.Println("cannot find leaderboard")
		}

		var wg sync.WaitGroup
		for _, track := range leaderboard.Tracks {
			for _, stage := range track.Stages {
				wg.Add(1)
				go func() {
					defer wg.Done()
					records, err := collectors.ExtractTimingTable(track.Name, stage)
					if err != nil {
						return
					}

					jsonRecords, err := json.Marshal(records)
					if err != nil {
						fmt.Println("error while marshaling json", err)
						return
					}

					err = s.Put(store.Record{
						Timestamp: uint32(time.Now().Unix()),
						Key:       []byte(stage.Name),
						Value:     jsonRecords,
					})
					if err != nil {
						return
					}
				}()
			}

		}
		wg.Wait()
	}
	return nil
}

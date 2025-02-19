package leaderboard

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dimfu/kaido/collectors"
	"github.com/dimfu/kaido/config"
	"github.com/dimfu/kaido/models"
	"github.com/dimfu/kaido/store"
	"github.com/urfave/cli/v3"
)

var cfg = config.GetConfig()

func List(ctx context.Context, c *cli.Command) error {
	for region := range cfg.Leaderboards {
		fmt.Println(region)
	}
	return nil
}

func Extract(ctx context.Context, c *cli.Command) error {
	cfg := config.GetConfig()
	s, err := store.GetInstance()
	if err != nil {
		return err
	}

	leaderboardFlag := c.String("leaderboard")
	currentMonth := c.Bool("current_month")

	leaderboard := strings.ToLower(leaderboardFlag)

	timing := collectors.TimingTable{
		Store:        s,
		Cfg:          cfg,
		CurrentMonth: currentMonth,
	}

	re := regexp.MustCompile(`\s*,\s*`)
	leaderboards := re.Split(leaderboard, -1)

	// handle if one of the leaderboard item including "all" by adding the whole leaderboard list instead
	for _, leaderboard := range leaderboards {
		if leaderboard == "all" {
			leaderboards = make([]string, 0, len(cfg.Leaderboards))
			for region := range cfg.Leaderboards {
				leaderboards = append(leaderboards, region)
			}
			break
		}
	}

	var wg sync.WaitGroup

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	for _, leaderboard := range leaderboards {
		wg.Add(1)
		go func(leaderboard string) {
			defer wg.Done()
			results, err := timing.Extract(leaderboard)
			if err != nil {
				fmt.Println(err)
				return
			}

			for region, result := range results {
				compare(region, *result.Prev, *result.Curr)
			}
		}(leaderboard)
	}

	select {
	case <-done:
		fmt.Println("Success collecting records from", len(leaderboards), "leaderboards")
	case <-time.After(10 * time.Second):
		fmt.Println("Timeout")
	}

	return nil
}

func compare(region string, prev, curr models.Record) {
	eq := reflect.DeepEqual(prev, curr)
	if !eq {
		fmt.Println(curr.Player, "is the new winner in", region, "by", curr.Time)
	}
}

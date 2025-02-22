package leaderboard

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dimfu/kaido/collectors"
	"github.com/dimfu/kaido/config"
	"github.com/dimfu/kaido/discord"
	"github.com/dimfu/kaido/models"
	"github.com/dimfu/kaido/store"
	"github.com/urfave/cli/v3"
)

var (
	cfg = config.GetConfig()
	mu  sync.Mutex
	sb  strings.Builder
)

func List(ctx context.Context, c *cli.Command) error {
	for region := range cfg.Leaderboards {
		fmt.Println(region)
	}
	return nil
}

func Extract(ctx context.Context, c *cli.Command) error {
	start := time.Now()
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

	defer func() {
	}()

	var wg sync.WaitGroup

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	for _, leaderboard := range leaderboards {
		wg.Add(1)
		go func(leaderboard string, currentMonth bool) {
			defer wg.Done()
			results, err := timing.Extract(leaderboard)
			if err != nil {
				fmt.Println(err)
				return
			}

			for region, result := range results {
				currWinnerStr, err := compare(region, result.Prev, result.Curr, currentMonth)
				if err != nil {
					fmt.Println(err)
				}
				mu.Lock()
				sb.Write([]byte(currWinnerStr))
				mu.Unlock()
			}
		}(leaderboard, currentMonth)
	}

	select {
	case <-done:
		if sb.Len() > 0 {
			err := discord.Send(sb.String(), cfg.DiscordWebhookURL)
			if err != nil {
				fmt.Println(err)
			}
		}

		fmt.Printf("Success collecting records from %d leaderboards (took %s)\n", len(leaderboards), time.Since(start))
	case <-time.After(10 * time.Second):
		fmt.Println("Timeout")
	}

	return nil
}

func toSeconds(time string) (float64, error) {
	parts := strings.Split(time, ":")

	if len(parts) != 2 {
		return -1, errors.New("record time is not valid, must be MM:SS.mmm")
	}

	minutes, err := strconv.Atoi(parts[0])
	if err != nil {
		return -1, err
	}

	secondParts := strings.Split(parts[1], ".")
	seconds, err := strconv.Atoi(secondParts[0])
	if err != nil {
		return -1, err
	}

	milliseconds := 0
	if len(secondParts) == 2 {
		milliseconds, err = strconv.Atoi(secondParts[1])
		if err != nil {
			return 0, err
		}
	}

	return float64(minutes*60+seconds) + float64(milliseconds)/1000.0, nil
}

func compare(region string, prev, curr *models.Record, currMonth bool) (string, error) {
	if len(prev.Time) == 0 {
		return "", errors.New("skipping comparison, previous record is not yet written in the store")
	}

	t1, err := toSeconds(prev.Time)
	if err != nil {
		return "", err
	}
	t2, err := toSeconds(curr.Time)
	if err != nil {
		return "", err
	}

	if t2 < t1 {
		if !currMonth { // all time
			return fmt.Sprintf("%s is the new fastest time in %s by %s\n", curr.Player, region, curr.Time), nil
		} else {
			return fmt.Sprintf("%s is the new fastest time this month in %s by %s\n", curr.Player, region, curr.Time), nil
		}
	}

	return "", nil
}

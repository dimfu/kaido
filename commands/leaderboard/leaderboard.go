package leaderboard

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
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

const (
	ALL_TIME_STR   = "New all-time fastest lap! %s in %s by %s\n"
	CURR_MONTH_STR = "New fastest lap this month! %s in %s by %s\n"
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
	if slices.Contains(leaderboards, "all") {
		leaderboards = make([]string, 0, len(cfg.Leaderboards))
		for region := range cfg.Leaderboards {
			leaderboards = append(leaderboards, region)
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
		go func(leaderboard string, currentMonth bool) {
			defer wg.Done()
			results, err := timing.Extract(leaderboard)
			if err != nil {
				fmt.Println(err)
				return
			}

			for region, result := range results {
				currTop, err := compare(region, result.Prev, result.Curr, currentMonth)
				if err != nil {
					fmt.Println(err)
				}
				mu.Lock()
				sb.Write([]byte(currTop))
				mu.Unlock()
			}
		}(leaderboard, currentMonth)
	}

	select {
	case <-done:
		if sb.Len() > 0 {
			players := strings.Split(sb.String(), "\n")
			batchSize := 10

			var wg sync.WaitGroup

			for i := 0; i < len(players); i += batchSize {
				wg.Add(1)
				end := i + batchSize
				if end > len(players) {
					end = len(players)
				}
				go func(ps []string) {
					defer wg.Done()
					toStr := strings.Join(ps, "\n")
					err := discord.Send(toStr, cfg.DiscordWebhookURL)
					if err != nil {
						fmt.Println(err)
					}
				}(players[i:end])
			}
			wg.Wait()
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

func compare(region string, prev, curr []models.Record, currMonth bool) (string, error) {
	prevFirst, currFirst := getFastestRecord(prev), getFastestRecord(curr)

	if prevFirst == nil && currFirst == nil {
		return "", fmt.Errorf("Cannot find records in %s, skipping...", region)
	}

	var t1, t2 float64
	var err error

	if prevFirst != nil {
		t1, err = toSeconds(prevFirst.Time)
		if err != nil {
			return "", err
		}
	}

	if currFirst != nil {
		t2, err = toSeconds(currFirst.Time)
		if err != nil {
			return "", err
		}

		// handle current month winner if there is no prev record
		if prevFirst == nil && currMonth {
			return fmt.Sprintf(CURR_MONTH_STR, currFirst.Time, region, currFirst.Player), nil
		}
	} else {
		return "", fmt.Errorf("Nothing to compare in %s leaderboard", region)
	}

	if t2 < t1 {
		msg := ALL_TIME_STR
		if currMonth {
			msg = CURR_MONTH_STR
		}
		return fmt.Sprintf(msg, currFirst.Time, region, currFirst.Player), nil
	}

	return "", nil
}

func getFastestRecord(records []models.Record) *models.Record {
	if len(records) > 0 && records[0].Rank == 1 {
		return &records[0]
	}
	return nil
}

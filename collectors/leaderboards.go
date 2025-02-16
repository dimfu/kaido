package collectors

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/dimfu/kaido/config"
	"github.com/dimfu/kaido/models"
	"github.com/gocolly/colly"
)

var (
	ERR_ALREADY_GENERATED = errors.New("leaderboard tracks already generated")
)

func GenerateLeaderboardTracks() error {
	cfg := config.GetConfig()

	if cfg.Leaderboards != nil {
		return ERR_ALREADY_GENERATED
	}

	leaderboards, err := getLeaderboards()
	if err != nil {
		return err
	}

	if cfg.Leaderboards == nil {
		cfg.Leaderboards = make(models.Leaderboards)
	}

	var wg sync.WaitGroup
	for region, leaderboard := range *leaderboards {
		wg.Add(1)
		cfg.Leaderboards[region] = leaderboard
		go func(u, r string) {
			defer wg.Done()
			tracks, err := getTracks(u)
			if err != nil {
				return
			}
			for _, track := range tracks {
				entry, exists := cfg.Leaderboards[r]
				if !exists {
					continue
				}
				entry.Tracks = append(entry.Tracks, *track)
				cfg.Leaderboards[region] = entry
			}
		}(leaderboard.Url, region)
	}
	wg.Wait()

	if err := cfg.Save(); err != nil {
		return err
	}

	return nil
}

func getLeaderboards() (*map[string]models.Leaderboard, error) {
	cfg := config.GetConfig()
	leaderboard := make(map[string]models.Leaderboard)
	c := colly.NewCollector()

	// build leaderboard map
	c.OnHTML("#Timing-navbar-dropdown", func(h *colly.HTMLElement) {
		next := h.DOM.Next()
		if next == nil {
			return
		}

		for _, node := range next.Children().Nodes {
			if node.Data != "a" {
				continue
			}
			elem := colly.NewHTMLElementFromSelectionNode(h.Response, next, node, 0)
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					u, err := url.Parse(fmt.Sprintf("%v%v", cfg.KBTBaseUrl, attr.Val))
					if err != nil {
						continue
					}

					region := strings.ToLower(elem.Text)

					if _, exists := u.Query()["leaderboard"]; exists {
						leaderboard[region] = models.Leaderboard{
							Region: region,
							Url:    u.String(),
						}
					}
				}
			}

		}
	})

	err := c.Visit(cfg.KBTBaseUrl)
	if err != nil {
		return nil, err
	}

	return &leaderboard, nil
}

func getTracks(regionUrl string) (map[string]*models.Track, error) {
	tracks := make(map[string]*models.Track)

	c := colly.NewCollector()

	c.OnHTML("select[name='track']", func(h *colly.HTMLElement) {
		for _, t := range strings.Fields(h.Text) {
			tracks[t] = &models.Track{
				Name: t,
			}
		}
	})

	if err := c.Visit(regionUrl); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	for key, track := range tracks {
		wg.Add(1)
		go func(track models.Track, key, r string) {
			defer wg.Done()
			stages, err := getStages(r, track.Name)
			if err != nil {
				return
			}
			for _, stage := range stages {
				stageUrl, err := buildStageUrl(r, track.Name, stage)
				if err != nil {
					continue
				}
				track.Stages = append(track.Stages, models.Stage{
					Name: stage,
					Url:  stageUrl,
				})
			}
			tracks[key] = &track
		}(*track, key, regionUrl)
	}
	wg.Wait()

	return tracks, nil
}

func getStages(regionUrl string, track string) ([]string, error) {
	var stages []string
	c := colly.NewCollector()

	c.OnHTML("select[name='stage']", func(h *colly.HTMLElement) {
		for _, node := range h.DOM.Children().Nodes {
			elem := colly.NewHTMLElementFromSelectionNode(h.Response, h.DOM, node, 0)
			stages = append(stages, elem.Text)
		}
	})

	u, err := url.Parse(regionUrl)
	if err != nil {
		return stages, err
	}
	q := u.Query()
	q.Set("track", track)
	u.RawQuery = q.Encode()
	if err := c.Visit(u.String()); err != nil {
		return stages, err
	}
	return stages, nil
}

func buildStageUrl(base, track, stage string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	q := u.Query()

	q.Set("track", track)
	q.Set("stage", stage)

	u.RawQuery = strings.ReplaceAll(q.Encode(), "&", "&")

	return u.String(), nil
}

package collectors

import (
	"errors"
	"fmt"
	"net/url"

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

	leaderboards, err := getLeaderboardTracks()
	if err != nil {
		return err
	}

	if cfg.Leaderboards == nil {
		cfg.Leaderboards = make(models.Leaderboards)
	}

	for region, leaderboard := range *leaderboards {
		cfg.Leaderboards[region] = leaderboard
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	return nil
}

func getLeaderboardTracks() (*map[string]models.Leaderboard, error) {
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

					if _, exists := u.Query()["leaderboard"]; exists {
						leaderboard[elem.Text] = models.Leaderboard{
							Region: elem.Text,
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

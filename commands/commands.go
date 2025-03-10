package commands

import (
	"context"

	"github.com/dimfu/kaido/commands/leaderboard"
	"github.com/dimfu/kaido/discord"
	"github.com/urfave/cli/v3"
)

func Commands() []*cli.Command {
	return []*cli.Command{
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
				&cli.BoolFlag{
					Name:    "current_month",
					Value:   false,
					Usage:   "scope result only for current month",
					Aliases: []string{"c"},
				},
			},
			Action: leaderboard.Extract,
		},
		{
			Name:   "leaderboards",
			Usage:  "See all available leaderboards",
			Action: leaderboard.List,
		},
		{
			Name:  "webhook",
			Usage: "options for webhook",
			Commands: []*cli.Command{
				{
					Name:  "set",
					Usage: "[web hook url] to update current webhook",
					Action: func(ctx context.Context, c *cli.Command) error {
						err := discord.Prompt(c.Args().First())
						if err != nil {
							return err
						}
						return nil
					},
				},
			},
		},
	}
}

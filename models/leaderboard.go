package models

type Track struct {
	Name   string   `json:"name"`
	Stages []string `json:"stages"`
	Url    string   `json:"url"`
}

type Leaderboard struct {
	Region string  `json:"region"`
	Tracks []Track `json:"tracks"`
	Url    string  `json:"url"`
}

type Leaderboards = map[string]Leaderboard

package models

type Track struct {
	Name   string  `json:"name"`
	Stages []Stage `json:"stages"`
}

type Stage struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type Leaderboard struct {
	Region string  `json:"region"`
	Tracks []Track `json:"tracks"`
	Url    string  `json:"url"`
}

type Leaderboards = map[string]Leaderboard

type TimingLeaderboard struct {
	Leaderboard Leaderboard `json:"leaderboard"`
	Records     []Record    `json:"records"`
}

type Record struct {
	Rank    int    `json:"rank"`
	Date    string `json:"date"`
	Player  string `json:"player"`
	CarName string `json:"carName"`
	Time    string `json:"time"`
}

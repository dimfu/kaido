package collectors

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/dimfu/kaido/config"
	"github.com/dimfu/kaido/models"
	"github.com/dimfu/kaido/store"
	"github.com/gocolly/colly"
)

type TimingTable struct {
	Store *store.Store
	Cfg   *config.Config
	wg    sync.WaitGroup
}

type TimingResult struct {
	Prev  *models.Record
	Curr  *models.Record
	stage string
	err   error
}

func (t *TimingTable) Extract(l string) (map[string]TimingResult, error) {
	var stages []models.Stage
	result := make(map[string]TimingResult)
	leaderboard, exists := t.Cfg.Leaderboards[l]
	if !exists {
		return nil, fmt.Errorf("cannot find leaderboard: %s", l)
	}

	for _, track := range leaderboard.Tracks {
		stages = append(stages, track.Stages...)
	}

	resChan := make(chan TimingResult, len(stages))
	t.wg.Add(len(stages))
	for _, stage := range stages {
		go t.processRecords(stage, resChan)
	}

	go func() {
		t.wg.Wait()
		close(resChan)
	}()

	for r := range resChan {
		if r.err != nil {
			// TODO: handle errors properly
			fmt.Println(r.err)
			continue
		}
		result[r.stage] = TimingResult{
			Prev: r.Prev,
			Curr: r.Curr,
		}
	}

	return result, nil
}

func (t *TimingTable) processRecords(stage models.Stage, ch chan<- TimingResult) {
	defer t.wg.Done()

	prev, err := t.prevTimingRecords(stage.Name)
	if err != nil {
		if err != store.ERR_KEY_NOT_FOUND {
			ch <- TimingResult{err: err}
			return
		}
	}

	curr, err := t.getRecords(stage)
	if err != nil {
		ch <- TimingResult{err: err}
		return
	}

	if err := t.updateTimingRecords(curr, stage.Name); err != nil {
		ch <- TimingResult{err: err}
		return
	}

	var prevFirst, currFirst models.Record
	if len(prev) > 0 && prev[0].Rank == 1 {
		prevFirst = prev[0]
	}

	if len(curr) > 0 && curr[0].Rank == 1 {
		currFirst = curr[0]
	}

	if prevFirst == (models.Record{}) && currFirst == (models.Record{}) {
		ch <- TimingResult{err: fmt.Errorf("Both prevFirst and currFirst are empty, skipping...")}
		return
	}

	ch <- TimingResult{
		stage: stage.Name,
		Prev:  &prevFirst,
		Curr:  &currFirst,
	}
}

func (t *TimingTable) prevTimingRecords(stage string) ([]models.Record, error) {
	var records []models.Record
	r, err := t.Store.Get(stage)
	if err != nil {
		return records, err
	}
	if err := json.Unmarshal(r.Value, &records); err != nil {
		return records, err
	}
	return records, nil
}

func (t *TimingTable) updateTimingRecords(records []models.Record, stage string) error {
	jsonRecords, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("error while marshaling json: %v", err)
	}
	err = t.Store.Put(store.Record{
		Timestamp: uint32(time.Now().Unix()),
		Key:       []byte(stage),
		Value:     jsonRecords,
	})
	if err != nil {
		return fmt.Errorf("error while updating key store: %v", err)
	}
	return nil
}

func (t *TimingTable) getRecords(stage models.Stage) ([]models.Record, error) {
	records := []models.Record{}
	c := colly.NewCollector()

	c.OnHTML("table", func(h *colly.HTMLElement) {
		h.ForEach("tbody > tr", func(i int, h *colly.HTMLElement) {
			rank, _ := strconv.Atoi(h.ChildText("td:nth-of-type(1)"))
			records = append(records, models.Record{
				Rank:    rank,
				Date:    h.ChildText("td:nth-of-type(2)"),
				Player:  h.ChildText("td:nth-of-type(3)"),
				CarName: h.ChildText("td:nth-of-type(4)"),
				Time:    h.ChildText("td:nth-of-type(5)"),
			})
		})
	})

	err := c.Visit(stage.Url)
	if err != nil {
		return records, err
	}

	return records, nil
}

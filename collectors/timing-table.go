package collectors

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/dimfu/kaido/config"
	"github.com/dimfu/kaido/models"
	"github.com/dimfu/kaido/store"
	"github.com/gocolly/colly"
)

type TimingTable struct {
	Store        *store.Store
	Cfg          *config.Config
	CurrentMonth bool
	wg           sync.WaitGroup
}

type TimingResult struct {
	Prev  []models.Record
	Curr  []models.Record
	stage string
	err   error
}

func (t *TimingTable) stageKey(trackName, stage string) string {
	if t.CurrentMonth {
		year, month, _ := time.Now().Date()
		return fmt.Sprintf("%d-%d_%s-%s", year, month, trackName, stage)
	} else {
		return stage
	}
}

func (t *TimingTable) Extract(l string) (map[string]TimingResult, error) {
	tracks := make(map[string][]models.Stage)
	result := make(map[string]TimingResult)
	leaderboard, exists := t.Cfg.Leaderboards[l]
	if !exists {
		return nil, fmt.Errorf("cannot find leaderboard: %s", l)
	}

	var stages int
	for _, track := range leaderboard.Tracks {
		tracks[track.Name] = append(tracks[track.Name], track.Stages...)
		stages += len(track.Stages)
	}

	resChan := make(chan TimingResult, stages)
	t.wg.Add(stages)
	for trackName, stages := range tracks {
		for _, stage := range stages {
			go t.processRecords(trackName, stage, resChan)
		}
	}

	go func() {
		t.wg.Wait()
		close(resChan)
	}()

	for r := range resChan {
		if r.err != nil {
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

func (t *TimingTable) processRecords(trackName string, stage models.Stage, ch chan<- TimingResult) {
	defer t.wg.Done()

	prev, err := t.prevTimingRecords(trackName, stage.Name)
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

	if err := t.updateTimingRecords(curr, trackName, stage.Name); err != nil {
		ch <- TimingResult{err: err}
		return
	}

	ch <- TimingResult{
		stage: stage.Name,
		Prev:  prev,
		Curr:  curr,
	}
}

func (t *TimingTable) prevTimingRecords(trackName, stage string) ([]models.Record, error) {
	var records []models.Record
	r, err := t.Store.Get(t.stageKey(trackName, stage))
	if err != nil {
		return records, err
	}
	if err := json.Unmarshal(r.Value, &records); err != nil {
		return records, err
	}
	return records, nil
}

func (t *TimingTable) updateTimingRecords(records []models.Record, trackName, stage string) error {
	jsonRecords, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("error while marshaling json: %v", err)
	}
	err = t.Store.Put(store.Record{
		Timestamp: uint32(time.Now().Unix()),
		Key:       []byte(t.stageKey(trackName, stage)),
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

	u, err := url.Parse(stage.Url)
	if err != nil {
		return records, err
	}

	q := u.Query()
	month := "0"
	if t.CurrentMonth {
		month = "1"
	}
	q.Set("month", month)
	u.RawQuery = q.Encode()

	err = c.Visit(u.String())
	if err != nil {
		return records, err
	}

	return records, nil
}

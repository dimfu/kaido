package collectors

import (
	"strconv"

	"github.com/dimfu/kaido/models"
	"github.com/gocolly/colly"
)

func ExtractTimingTable(region string, stage models.Stage) ([]models.Record, error) {
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

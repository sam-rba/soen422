package main

import (
	"fmt"
	"github.com/wcharczuk/go-chart/v2"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"
)

type ChartHandler struct {
	building Building
}

func (h ChartHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL)

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "invalid method: '%s'", r.Method)
		return
	}

	graph := chart.Chart{
		Background: chart.Style{
			Padding: chart.Box{Top: 20, Left: 20},
		},
		Series: buildSortedSeries(h.building),
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	w.Header().Set("Content-Type", "image/png")
	if err := graph.Render(chart.PNG, w); err != nil {
		log.Println(err)
	}
}

// Build a time series for each humidity record in the building.
// The returned series' are sorted by room name.
func buildSortedSeries(building Building) []chart.Series {
	seriesChan := make(chan chart.TimeSeries)
	sortedSeriesChan := make(chan []chart.Series)
	go sortSeriesByName(seriesChan, sortedSeriesChan)

	var wg sync.WaitGroup
	for room, record := range building {
		entries := make(chan Entry[Humidity])
		record.get <- entries
		wg.Add(1)
		go func() {
			buildSeries(room, entries, seriesChan)
			wg.Done()
		}()
	}
	wg.Wait()
	close(seriesChan)
	return <-sortedSeriesChan
}

// Sort a sequence of series' by name.
func sortSeriesByName(in <-chan chart.TimeSeries, out chan<- []chart.Series) {
	series := make([]chart.Series, 0)
	cmp := func(a chart.Series, b chart.TimeSeries) int {
		return strings.Compare(a.GetName(), b.GetName())
	}

	// Insertion sort.
	for s := range in {
		i, _ := slices.BinarySearchFunc(series, s, cmp)
		series = slices.Insert(series, i, chart.Series(s))
	}

	out <- series
	close(out)
}

// Build a time series from a sequence of record entries.
func buildSeries(room RoomID, in <-chan Entry[Humidity], out chan<- chart.TimeSeries) {
	var x []time.Time
	var y []float64

	for e := range in {
		x = append(x, e.t)
		y = append(y, float64(e.v))
	}
	out <- chart.TimeSeries{
		Name:    string(room),
		XValues: x,
		YValues: y,
	}
}

package main

import (
	"log"
	"net/http"
	"fmt"
	"time"
	"github.com/wcharczuk/go-chart/v2"
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

	var series []chart.Series
	for room, record := range h.building {
		var x []time.Time
		var y []float64
		c := make(chan Entry[Humidity])
		record.get <- c
		for e := range c {
			x = append(x, e.t)
			y = append(y, float64(e.v))
		}
		series = append(series, chart.TimeSeries{
			Name: string(room),
			XValues: x,
			YValues: y,
		})
	}

	graph := chart.Chart{
		Background: chart.Style{
			Padding: chart.Box{Top: 20, Left: 20},
		},
		Series: series,
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	w.Header().Set("Content-Type", "image/png")
	if err := graph.Render(chart.PNG, w); err != nil {
		log.Println(err)
	}
}

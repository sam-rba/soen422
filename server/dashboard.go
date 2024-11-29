package main

import (
	_ "embed"
	"fmt"
	"github.com/sam-rba/share"
	"html/template"
	"log"
	"net/http"
)

//go:embed dashboard.html
var dashboardHtml string

var dashboard = template.Must(template.New("dashboard").Parse(dashboardHtml))

type Dashboard struct {
	Target    Humidity
	Average   Humidity
	DutyCycle DutyCycle
	Rooms     map[RoomID]Humidity
}

type DashboardHandler struct {
	target    share.Val[Humidity]
	building  Building
	dutyCycle Record[DutyCycle]
}

func (h DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL)

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "invalid method: '%s'", r.Method)
		return
	}

	db := h.buildDashboard()
	err := dashboard.Execute(w, db)
	if err != nil {
		log.Println(err)
	}
}

func (h DashboardHandler) buildDashboard() Dashboard {
	var target Humidity
	if targetp, ok := h.target.TryGet(); ok {
		target = *targetp
	} else {
		target = 0
	}

	average, ok := h.building.average()
	if !ok {
		average = -1
	}

	c := make(chan Entry[DutyCycle])
	h.dutyCycle.getRecent <- c
	var duty DutyCycle
	if e, ok := <-c; ok {
		duty = e.v
	} else  {
		duty = -1
	}

	rooms := make(map[RoomID]Humidity)
	for id, record := range h.building {
		c := make(chan Entry[Humidity])
		record.getRecent <- c
		var humidity Humidity
		if e, ok := <-c; ok {
			humidity = e.v
		} else {
			humidity = -1
		}
		rooms[id] = humidity
	}

	return Dashboard{target, average, duty, rooms}
}

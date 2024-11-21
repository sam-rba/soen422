package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

const dashboardHtml = `
<!DOCTYPE html>
<html>
	<head>
		<title>HVAC Dashboard</title>
	</head>
	<body>
		<p>Average humidity: {{ printf "%.1f %%" .Average }}</p>
		<table>
			<tr><th>Room</th><th>Humidity</th></tr>
			{{ range .Rooms }}
				<tr><td>{{ .RoomID }}</td><td>{{ printf "%.1f %%" .Humidity }}</td></tr>
			{{ end }}
		</table>
	</body>
</html>`

var dashboard = template.Must(template.New("dashboard").Parse(dashboardHtml))

type Dashboard struct {
	Average Humidity
	Rooms   []Room
}

type Room struct {
	RoomID
	Humidity
}

type DashboardHandler struct {
	Building
}

func (h DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL)

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "invalid method: '%s'", r.Method)
		return
	}

	db := newDashboard(h.Building)
	err := dashboard.Execute(w, db)
	if err != nil {
		log.Println(err)
	}
}

func newDashboard(b Building) Dashboard {
	average, ok := b.average()
	if !ok {
		average = -1
	}

	// TODO: sort by room ID.
	rooms := make([]Room, 0, len(b))
	for id, record := range b {
		c := make(chan Humidity)
		record.getRecent <- c
		humidity, ok := <-c
		if !ok {
			humidity = -1
		}
		rooms = append(rooms, Room{id, humidity})
	}

	return Dashboard{average, rooms}
}

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
		<p>Average humidity:
			{{/* A value less than 0 means no data. */}}
			{{- if ge .Average 0.0 -}}
				{{ printf "%.1f%%" .Average }}
			{{- else -}}
				unknown
			{{- end }}</p>
		<table>
			<tr><th>Room</th><th>Humidity</th></tr>
			{{ range $id, $humidity := .Rooms }}
				<tr>
					<td>{{ $id }}</td>
					<td>
						{{/* A value less than 0 means no data. */}}
						{{- if ge $humidity 0.0 -}}
							{{ printf "%.1f%%" $humidity }}
						{{- else -}}
							unknown
						{{- end -}}
					</td>
				</tr>
			{{ end }}
		</table>
	</body>
</html>`

var dashboard = template.Must(template.New("dashboard").Parse(dashboardHtml))

type Dashboard struct {
	Average Humidity
	Rooms   map[RoomID]Humidity
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

	rooms := make(map[RoomID]Humidity)
	for id, record := range b {
		c := make(chan Humidity)
		record.getRecent <- c
		humidity, ok := <-c
		if !ok {
			humidity = -1
		}
		rooms[id] = humidity
	}

	return Dashboard{average, rooms}
}

package main

import (
	"fmt"
	"github.com/sam-rba/share"
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
			{{- end -}}
		</p>
		<p>Duty cycle:
			{{/* A value less than 0 means no data. */}}
			{{- if ge .DutyCycle 0.0 -}}
				{{ printf "%.1f%%" .DutyCycle }}
			{{- else -}}
				unknown
			{{- end -}}
		</p>
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
	Average   Humidity
	DutyCycle DutyCycle
	Rooms     map[RoomID]Humidity
}

type DashboardHandler struct {
	building  Building
	dutyCycle share.Val[DutyCycle]
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
	average, ok := h.building.average()
	if !ok {
		average = -1
	}

	var duty DutyCycle
	if dutyp, ok := h.dutyCycle.TryGet(); ok {
		duty = *dutyp
	} else {
		duty = -1
	}

	rooms := make(map[RoomID]Humidity)
	for id, record := range h.building {
		c := make(chan Humidity)
		record.getRecent <- c
		humidity, ok := <-c
		if !ok {
			humidity = -1
		}
		rooms[id] = humidity
	}

	return Dashboard{average, duty, rooms}
}

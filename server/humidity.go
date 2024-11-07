package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Humidity float32

type HumidityHandler struct {
	rooms map[RoomID]Record[Humidity]
}

func newHumidityHandler(rooms []RoomID) HumidityHandler {
	h := HumidityHandler{make(map[RoomID]Record[Humidity])}
	for _, id := range rooms {
		h.rooms[id] = newRecord[Humidity]()
	}
	return h
}

func (h HumidityHandler) Close() {
	for _, record := range h.rooms {
		record.Close()
	}
}

func (h HumidityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL)
	switch r.Method {
	case http.MethodGet:
		h.get(w, r)
	case http.MethodPost:
		h.post(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "invalid method: '%s'", r.Method)
	}
}

func (h HumidityHandler) get(w http.ResponseWriter, r *http.Request) {
	if humidity, ok := h.average(); ok {
		fmt.Fprintf(w, "%.2f", humidity)
	} else {
		w.WriteHeader(http.StatusGone)
		fmt.Fprintf(w, "no humidity data stored on server")
	}
}

func (h HumidityHandler) post(w http.ResponseWriter, r *http.Request) {
	queryVals, err := parseQuery(r.URL.RawQuery, []string{"room", "humidity"})
	if err != nil {
		badRequest(w, "invalid query: %v", err)
		return
	}
	room := RoomID(queryVals["room"])
	humidityStr := queryVals["humidity"]

	humidity, err := strconv.ParseFloat(humidityStr, 32)
	if err != nil {
		badRequest(w, "invalid humidity: '%s'", humidityStr)
		return
	}

	record, ok := h.rooms[room]
	if !ok {
		badRequest(w, "invalid room ID: '%s'", room)
		return
	}

	record.put <- Humidity(humidity)
}

// Calculate the average humidity in the building. Returns false if there is not enough data available.
func (h HumidityHandler) average() (Humidity, bool) {
	var sum Humidity = 0
	nRooms := 0
	for room, record := range h.rooms {
		c := make(chan Humidity)
		record.getRecent <- c
		if humidity, ok := <-c; ok {
			sum += humidity
			nRooms++
		} else {
			log.Printf("Warning: no humidity for room '%s'\n", room)
		}
	}
	if nRooms == 0 {
		log.Println("Warning: not enough data to calculate average humidity")
		return -1.0, false
	}
	return sum / Humidity(nRooms), true
}

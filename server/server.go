package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

const addr = ":9090"
var rooms = []RoomID {
	",4AL[+V*:*k*n{7vL{}/d=K#Mo*y*^.@",
	"Jq!+<p3g-iu%-vU]FZp2H,AKZWp@!4![",
}

type Humidity float32
type RoomID string

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
	c := make(chan float32)
	h.humidity.getRecent <- c
	humidity, ok := <-c
	if !ok {
		w.WriteHeader(http.StatusGone)
		fmt.Fprintf(w, "no humidity data stored on server")
		return
	}
	fmt.Fprintf(w, "%.2f", humidity)
}

func (h HumidityHandler) post(w http.ResponseWriter, r *http.Request) {
	query := r.URL.RawQuery
	humidity, err := strconv.ParseFloat(query, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid query string: '%s'", query)
		return
	}
	h.humidity.put <- float32(humidity)
}

func main() {
	humidityHandler := newHumidityHandler()
	defer humidityHandler.Close()

	http.Handle("/humidity", humidityHandler)
	fmt.Printf("Listening on %s...\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

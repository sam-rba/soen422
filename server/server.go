package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

const addr = ":9090"

var rooms = []RoomID{
	"SNbeEcs7XVWMEvjeEYgwZnp9XYjToVhh",
	"rEKKa5TW5xjArmR25wT4Uiw7tksk4noE",
}

type Humidity float32
type RoomID string

type HumidityHandler struct {
	rooms map[RoomID]Record[Humidity]
}

type TargetHumidityHandler struct {
	mu     sync.Mutex
	target Humidity
}

func main() {
	humidityHandler := newHumidityHandler(rooms)
	defer humidityHandler.Close()

	http.Handle("/humidity", humidityHandler)
	http.Handle("/target_humidity", new(TargetHumidityHandler))
	fmt.Printf("Listening on %s...\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
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
		log.Println(err)
		badRequest(w, "invalid query: %v", err)
		return
	}
	room := RoomID(queryVals["room"])
	humidityStr := queryVals["humidity"]

	humidity, err := strconv.ParseFloat(humidityStr, 32)
	if err != nil {
		log.Println("Warning: invalid humidity:", err)
		badRequest(w, "invalid humidity: '%s'", humidityStr)
		return
	}

	record, ok := h.rooms[room]
	if !ok {
		log.Println("Warning: invalid room:", room)
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

func (h *TargetHumidityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h *TargetHumidityHandler) get(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	fmt.Fprintf(w, "%.2f", h.target)
}

func (h *TargetHumidityHandler) post(w http.ResponseWriter, r *http.Request) {
	target, err := strconv.ParseFloat(r.URL.RawQuery, 32)
	if err != nil {
		badRequest(w, "invalid humidity: '%s'", r.URL.RawQuery)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	h.target = Humidity(target)
}

// Parse the value associated with each key in the query string. Returns a map of
// keys and values, or error if one of the keys is missing, or if there is no value
// associated with one of the keys.
func parseQuery(query string, keys []string) (map[string]string, error) {
	queryVals, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}

	vals := make(map[string]string)
	for _, key := range keys {
		val := queryVals.Get(key)
		if val == "" {
			return nil, fmt.Errorf("missing key '%s'", key)
		}
		vals[key] = val
	}
	return vals, nil
}

func badRequest(w http.ResponseWriter, format string, a ...any) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, format, a)
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const (
	minHumidity = 0.0
	maxHumidity = 100.0
)

type Humidity float32

type HumidityHandler struct {
	Building
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
	if humidity, ok := h.Building.average(); ok {
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
	if err != nil || !isValidHumidity(humidity){
		badRequest(w, "invalid humidity: '%s'", humidityStr)
		return
	}

	record, ok := h.Building[room]
	if !ok {
		badRequest(w, "invalid room ID: '%s'", room)
		return
	}

	record.put <- Humidity(humidity)
}

// Parse the value associated with each key in the query string. Returns a map of
// keys and values, or error if one of the keys is missing or if there is no value
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

func isValidHumidity(humidity float64) bool {
	return humidity >= minHumidity && humidity <= maxHumidity;
}

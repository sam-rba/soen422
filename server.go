package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

const addr = ":9090"

type humidityHandler struct {
	humidity Record[float32]
}

func newHumidityHandler() humidityHandler {
	return humidityHandler{newRecord[float32]()}
}

func (h humidityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h humidityHandler) get(w http.ResponseWriter, r *http.Request) {
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

func (h humidityHandler) post(w http.ResponseWriter, r *http.Request) {
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
	http.Handle("/humidity", newHumidityHandler())
	fmt.Printf("Listening on %s...\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type TargetHumidityHandler struct {
	mu     sync.Mutex
	target Humidity
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

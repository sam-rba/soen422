package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type DutyCycle float32

type DutyCycleHandler struct {
	mu sync.Mutex
	dc DutyCycle
}

func (h *DutyCycleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL)

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "invalid method: '%s'", r.Method)
		return
	}

	dc, err := strconv.ParseFloat(r.URL.RawQuery, 32)
	if err != nil {
		badRequest(w, "invalid duty cycle: '%s'", r.URL.RawQuery)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	h.dc = DutyCycle(dc)
}

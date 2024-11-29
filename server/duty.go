package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

const (
	minDutyCycle = 0.0
	maxDutyCycle = 100.0
)

type DutyCycle float32

type DutyCycleHandler struct {
	dc Record[DutyCycle]
}

func (h DutyCycleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL)

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "invalid method: '%s'", r.Method)
		return
	}

	dc, err := strconv.ParseFloat(r.URL.RawQuery, 32)
	if err != nil || !isValidDutyCycle(dc) {
		badRequest(w, "invalid duty cycle: '%s'", r.URL.RawQuery)
		return
	}

	h.dc.put <- DutyCycle(dc)
}

func isValidDutyCycle(dc float64) bool {
	return dc >= minDutyCycle && dc <= maxDutyCycle
}

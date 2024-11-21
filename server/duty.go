package main

import (
	"fmt"
	"github.com/sam-rba/share"
	"log"
	"net/http"
	"strconv"
)

type DutyCycle float32

type DutyCycleHandler struct {
	dc share.Val[DutyCycle]
}

func (h DutyCycleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	h.dc.Set <- DutyCycle(dc)
}

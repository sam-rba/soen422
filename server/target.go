package main

import (
	"fmt"
	"github.com/sam-rba/share"
	"log"
	"net/http"
	"strconv"
)

type TargetHumidityHandler struct {
	target share.Val[Humidity]
}

func (h TargetHumidityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h TargetHumidityHandler) get(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%.2f", h.target.Get())
}

func (h TargetHumidityHandler) post(w http.ResponseWriter, r *http.Request) {
	target, err := strconv.ParseFloat(r.URL.RawQuery, 32)
	if err != nil {
		badRequest(w, "invalid humidity: '%s'", r.URL.RawQuery)
		return
	}

	h.target.Set <- Humidity(target)
}

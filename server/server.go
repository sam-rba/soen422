package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
)

const addr = ":9090"

var rooms = []RoomID{
	"SNbeEcs7XVWMEvjeEYgwZnp9XYjToVhh",
	"rEKKa5TW5xjArmR25wT4Uiw7tksk4noE",
}

type RoomID string

func main() {
	humidityHandler := newHumidityHandler(rooms)
	defer humidityHandler.Close()

	http.Handle("/humidity", humidityHandler)
	http.Handle("/target_humidity", new(TargetHumidityHandler))
	http.Handle("/duty_cycle", new(DutyCycleHandler))

	fmt.Printf("Listening on %s...\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
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
	log.Println("Warning: bad request:", fmt.Sprintf(format, a))
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, format, a)
}

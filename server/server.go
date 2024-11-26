package main

import (
	"fmt"
	"github.com/sam-rba/share"
	"log"
	"net/http"
)

const (
	addr = ":9090"

	targetHumidityDefault = 35.0
)

var roomIDs = []RoomID{
	"SNbeEcs7XVWMEvjeEYgwZnp9XYjToVhh",
	"rEKKa5TW5xjArmR25wT4Uiw7tksk4noE",
}

type RoomID string

func main() {
	target := share.NewVal[Humidity]()
	target.Set <- targetHumidityDefault
	building := newBuilding(roomIDs)
	dutyCycle := share.NewVal[DutyCycle]()
	defer target.Close()
	defer building.Close()
	defer dutyCycle.Close()

	http.Handle("/", DashboardHandler{target, building, dutyCycle})
	http.Handle("/humidity", HumidityHandler{building})
	http.Handle("/target_humidity", TargetHumidityHandler{target})
	http.Handle("/duty_cycle", DutyCycleHandler{dutyCycle})

	fmt.Printf("Listening on %s...\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func badRequest(w http.ResponseWriter, format string, a ...any) {
	log.Println("Warning: bad request:", fmt.Sprintf(format, a))
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, format, a)
}

package main

import "log"

type Building map[RoomID]Record[Humidity]

func newBuilding(roomIDs []RoomID) Building {
	b := make(Building)
	for _, id := range roomIDs {
		b[id] = newRecord[Humidity]()
	}
	return b
}

func (b Building) Close() {
	for _, record := range b {
		record.Close()
	}
}

// Calculate the average humidity in the building. Returns false if there is not enough data available.
func (b Building) average() (Humidity, bool) {
	var sum Humidity = 0
	nRooms := 0
	for room, record := range b {
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

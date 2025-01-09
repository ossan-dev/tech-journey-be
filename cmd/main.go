package main

import (
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	// get auth token
	httpClient := http.Client{}
	token, err := getAuthToken(&httpClient)
	if err != nil {
		panic(err)
	}
	// get rooms IDs
	roomsIDs := []string{}
	if err := getRoomsIDs(&roomsIDs); err != nil {
		panic(err)
	}
	// define how many bookings do you want to have per room
	bookings := make(chan string, 30000)
	bookingsToMake := make(map[string]int, 3)
	bookingsToMake[roomsIDs[0]] = 10000
	bookingsToMake[roomsIDs[1]] = 10000
	bookingsToMake[roomsIDs[2]] = 10000
	prepareBookingsToMake(bookingsToMake, bookings)

	// batch bookings creation
	batchBookingCreation(&httpClient, token, bookings)

	// understand how to create profiles
	// understand how to read profiles' results
}

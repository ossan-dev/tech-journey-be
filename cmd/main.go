package main

import (
	"fmt"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

const numberOfBookings = 30000

func main() {
	httpClient := http.Client{}

	// defer call to get FlightRecorder trace
	defer func() {
		res, err := httpClient.Get("http://localhost:8080/trace")
		if err != nil {
			panic(err)
		}
		if res.StatusCode != http.StatusOK {
			fmt.Fprintln(os.Stderr, "HTTP Server /trace endpoint responded with StatusCode:", res.StatusCode)
		}
	}()

	// get auth token
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
	bookings := make(chan string, numberOfBookings)
	bookingsToMake := make(map[string]int, 3)
	bookingsToMake[roomsIDs[0]] = numberOfBookings / 3
	bookingsToMake[roomsIDs[1]] = numberOfBookings / 3
	bookingsToMake[roomsIDs[2]] = numberOfBookings / 3
	prepareBookingsToMake(bookingsToMake, bookings)

	// batch bookings creation
	batchBookingCreation(&httpClient, token, bookings)
}

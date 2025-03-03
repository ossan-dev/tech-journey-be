package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var pool = &sync.Pool{
	New: func() any {
		r, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "http://localhost:8080/bookings", nil)
		r.Header.Set("Content-Type", "application/json")
		return r
	},
}

func prepareBookingsToMake(bookingsToMake map[string]int, bookings chan<- string) {
	for k, v := range bookingsToMake {
		for range v {
			bookings <- k
		}
	}
	close(bookings)
}

func batchBookingCreation(client *http.Client, token string, bookings <-chan string) {
	start := time.Now()
	defer func() {
		fmt.Printf("Elapsed time: %v\n", time.Since(start)) // ~51s
	}()
	var wg sync.WaitGroup
	wg.Add(runtime.NumCPU())
	for w := 1; w <= runtime.NumCPU(); w++ {
		go worker(client, token, bookings, &wg)
	}
	wg.Wait()
	fmt.Printf("number of workers: %v\n", runtime.NumCPU())
}

func worker(client *http.Client, token string, bookings <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for booking := range bookings {
		if err := addBooking(client, token, booking); err != nil {
			fmt.Printf("error while creating a booking: %v\n", err)
			os.Exit(1)
		}
	}
}

func addBooking(client *http.Client, token string, roomID string) (err error) {
	r := pool.Get().(*http.Request)
	r.Body = io.NopCloser(strings.NewReader(fmt.Sprintf(`{"room_id":"%v","booked_on":"2025-01-09"}`, roomID)))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	res, err := client.Do(r)
	if err != nil {
		return
	}
	defer res.Body.Close()
	pool.Put(r)
	if res.StatusCode != http.StatusCreated {
		err = fmt.Errorf("booking not created")
		return
	}
	responseTokens := make(map[string]any, 1)
	if err = json.NewDecoder(res.Body).Decode(&responseTokens); err != nil {
		err = fmt.Errorf("failed to decode http response")
		return
	}
	fmt.Printf("room id: %v, booking id: %v created!\n", roomID, responseTokens["id"])
	return
}

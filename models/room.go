package models

type Photo struct {
	ID     int64
	Url    string
	RoomId string
	Room   Room
}

type Room struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Cost          float64   `json:"cost"`
	NumberOfSeats int       `json:"number_of_seats"`
	Category      string    `json:"category"`
	MainPhoto     string    `json:"main_photo"`
	IsAvailable   bool      `gorm:"-:all"`
	Photos        []Photo   `json:"-"`
	Bookings      []Booking `json:"-"`
}

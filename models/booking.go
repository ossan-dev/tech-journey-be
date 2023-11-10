package models

import "time"

type Booking struct {
	ID        string    `json:"id"`
	BookedOn  time.Time `json:"booked_on"`
	CreatedAt time.Time `json:"created_at"`
	RoomId    string    `json:"room_id"`
	UserId    string    `json:"-"`
	Room      Room      `json:"-"`
	User      User      `json:"-"`
}

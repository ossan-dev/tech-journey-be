package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

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

func GetRooms(db *gorm.DB, dayToBook time.Time) (res []Room, err error) {
	err = db.Model(&Room{}).Preload("Bookings").Find(&res).Error
	if err != nil {
		return nil, err
	}
	for k, room := range res {
		res[k].IsAvailable = true
		for _, booking := range room.Bookings {
			if booking.BookedOn.Equal(dayToBook) {
				res[k].IsAvailable = false
				break
			}
		}
	}
	return
}

func GetRoomById(db *gorm.DB, id string) (res *Room, err error) {
	err = db.Model(&Room{}).First(&res, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return
}

func GetRoomPhotos(db *gorm.DB, id string) (res []string, err error) {
	_, err = GetRoomById(db, id)
	if err != nil {
		return nil, err
	}
	err = db.Model(&Photo{}).Where("room_id = ?", id).Select("url").Find(&res).Error
	if err != nil {
		return nil, err
	}
	return
}

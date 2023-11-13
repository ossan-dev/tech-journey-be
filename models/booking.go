package models

import (
	"errors"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type Booking struct {
	ID        string    `json:"id"`
	BookedOn  time.Time `json:"booked_on"`
	CreatedAt time.Time `json:"created_at"`
	RoomId    string    `json:"room_id"`
	UserId    string    `json:"-"`
	Room      Room      `json:"-"`
	User      User      `json:"-"`
}

func CreateBooking(db *gorm.DB, booking Booking) (*string, error) {
	err := db.Model(&Booking{}).Create(&booking).Error
	if err != nil {
		return nil, CoworkingErr{StatusCode: http.StatusInternalServerError, Code: DbErr, Message: err.Error()}
	}
	return &booking.ID, nil
}

func GetBookingsByUserId(db *gorm.DB, userId string) (res []Booking, err error) {
	err = db.Model(&Booking{}).Where("user_id = ?", userId).Find(&res).Error
	if err != nil {
		return nil, CoworkingErr{StatusCode: http.StatusInternalServerError, Code: DbErr, Message: err.Error()}
	}
	return
}

func GetBookingById(db *gorm.DB, id, userId string) (res *Booking, err error) {
	err = db.Model(&Booking{}).Where("id = ? and user_id = ?", id, userId).First(&res).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, CoworkingErr{StatusCode: http.StatusNotFound, Code: ObjectNotFoundErr, Message: err.Error()}
		}
		return nil, CoworkingErr{StatusCode: http.StatusInternalServerError, Code: DbErr, Message: err.Error()}
	}
	return
}

func DeleteBookingById(db *gorm.DB, id, userId string) error {
	booking, err := GetBookingById(db, id, userId)
	if err != nil {
		return err
	}
	if err := db.Model(&Booking{}).Delete(&booking).Error; err != nil {
		return CoworkingErr{StatusCode: http.StatusInternalServerError, Code: DbErr, Message: err.Error()}
	}
	return nil
}

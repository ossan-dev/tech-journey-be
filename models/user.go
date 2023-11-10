package models

import (
	"errors"
	"fmt"

	"coworkingapp/utils"

	"gorm.io/gorm"
)

type User struct {
	ID       string
	Email    string
	Username string
	Password string
}

func LoginUser(db *gorm.DB, username, password string) (res *User, err error) {
	if err = db.Model(&User{}).Where("username = ? and password = ?", username, password).First(&res).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return
}

func GetUserByEmail(db *gorm.DB, email string) (res *User, err error) {
	if err = db.Model(&User{}).Where("email = ?", email).First(&res).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return
}

func SignupUser(db *gorm.DB, user User) (id string, err error) {
	if err = db.Model(&User{}).First(&User{}, "email = ?", user.Email).Error; err == nil {
		return "", fmt.Errorf("email already registered")
	}
	user.ID = utils.GetUuid()
	if err = db.Model(&User{}).Create(&user).Error; err != nil {
		return "", err
	}
	return user.ID, nil
}

package models

import (
	"errors"
	"net/http"

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
			return nil, CoworkingErr{StatusCode: http.StatusNotFound, Code: InvalidCredentialsErr, Message: err.Error()}
		}
		return nil, CoworkingErr{StatusCode: http.StatusInternalServerError, Code: DbErr, Message: err.Error()}
	}
	return
}

func GetUserByEmail(db *gorm.DB, email string) (res *User, err error) {
	if err = db.Model(&User{}).Where("email = ?", email).First(&res).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, CoworkingErr{StatusCode: http.StatusNotFound, Code: ObjectNotFoundErr, Message: err.Error()}
		}
		return nil, CoworkingErr{StatusCode: http.StatusInternalServerError, Code: DbErr, Message: err.Error()}
	}
	return
}

func SignupUser(db *gorm.DB, user User) (id string, err error) {
	if err = db.Model(&User{}).First(&User{}, "email = ?", user.Email).Error; err == nil {
		return "", CoworkingErr{StatusCode: http.StatusBadRequest, Code: EmailAlreadyInUseErr, Message: "please change the email and retry"}
	}
	user.ID = utils.GetUuid()
	if err = db.Model(&User{}).Create(&user).Error; err != nil {
		return "", CoworkingErr{StatusCode: http.StatusInternalServerError, Code: DbErr, Message: err.Error()}
	}
	return user.ID, nil
}

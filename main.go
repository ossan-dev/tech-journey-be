package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ossan-dev/coworkingapp/handlers"
	"github.com/ossan-dev/coworkingapp/middlewares"
	"github.com/ossan-dev/coworkingapp/models"
	"github.com/ossan-dev/coworkingapp/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	var config models.CoworkingConfig
	data, err := os.ReadFile("config/config.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		panic(err)
	}
	if len(config.SecretKey) != 32 {
		panic(fmt.Errorf("config.SecretKey must have 32 bytes"))
	}
	pgConfig := postgres.Config{
		DSN: config.Dsn,
	}
	db, err := gorm.Open(postgres.Dialector{
		Config: &pgConfig,
	})
	if err != nil {
		panic(err)
	}
	user := models.User{}
	room := models.Room{}
	photo := models.Photo{}
	booking := models.Booking{}
	err = db.AutoMigrate(&user, &room, &photo, &booking)
	if err != nil {
		panic(err)
	}
	seedData(db)
	r := gin.Default()
	r.Use(middlewares.EarlyExitOnPreflightRequests())
	r.Use(middlewares.SetCorsPolicy(config.AllowedOrigins))
	r.Use(middlewares.SetRequestValues(*db, config))
	handlers.SetupRoutes(r)

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}

func seedData(db *gorm.DB) {
	sqldb, err := db.DB()
	if err != nil {
		panic(err)
	}
	defer sqldb.Close()

	_, err = sqldb.Exec(`DELETE FROM public.bookings`)
	if err != nil {
		panic(err)
	}

	_, err = sqldb.Exec(`DELETE FROM public.users`)
	if err != nil {
		panic(err)
	}

	_, err = sqldb.Exec(`DELETE FROM public.photos`)
	if err != nil {
		panic(err)
	}

	_, err = sqldb.Exec(`DELETE FROM public.rooms`)
	if err != nil {
		panic(err)
	}

	userId := utils.GetUuid()
	user := models.User{
		ID:       userId,
		Email:    "ipesenti@sorint.com",
		Username: "ipesenti",
		Password: "abcd1234!!",
	}
	db.Create(&user)
	photosRoomGreen := [2]models.Photo{
		{Url: "/green_0002.jpg"},
		{Url: "/green_0003.jpg"},
	}
	photosRoomRed := [1]models.Photo{
		{Url: "/red_0002.jpg"},
	}
	photosRoomYellow := [3]models.Photo{
		{Url: "/yellow_0002.jpg"},
		{Url: "/yellow_0003.jpg"},
		{Url: "/yellow_0004.jpg"},
	}
	rooms := [3]models.Room{
		{
			ID: utils.GetUuid(), Name: "Green", Cost: 12.50, NumberOfSeats: 4, Category: "Open Space", MainPhoto: "/green_0001.jpg", Photos: photosRoomGreen[:],
		},
		{
			ID: utils.GetUuid(), Name: "Red", Cost: 100.00, NumberOfSeats: 50, Category: "Conference Hall", MainPhoto: "/red_0001.jpg", Photos: photosRoomRed[:],
		},
		{
			ID: utils.GetUuid(), Name: "Yellow", Cost: 4.50, NumberOfSeats: 1, Category: "Shared Desk", MainPhoto: "/yellow_0001.jpg", Photos: photosRoomYellow[:],
		},
	}
	db.Create(&rooms)
}

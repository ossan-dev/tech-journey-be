package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gin-contrib/pprof"
	"github.com/google/gops/agent"
	"github.com/ossan-dev/coworkingapp/handlers"
	"github.com/ossan-dev/coworkingapp/middlewares"
	"github.com/ossan-dev/coworkingapp/models"
	"github.com/ossan-dev/coworkingapp/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// gops agent
	if err := agent.Listen(agent.Options{}); err != nil {
		panic(err)
	}

	// flight recorder
	fr := handlers.NewFlightRecorder()
	if err := fr.FlightRecorderTracer.Start(); err != nil {
		panic(err)
	}
	fr.FlightRecorderTracer.SetSize(4096)
	defer func() {
		if err := fr.FlightRecorderTracer.Stop(); err != nil {
			panic(err)
		}
	}()

	// app Main code
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
	utils.PrintMemStats()
	err = db.AutoMigrate(&user, &room, &photo, &booking)
	if err != nil {
		panic(err)
	}
	utils.PrintMemStats()
	seedData(db)
	r := gin.Default()
	r.Use(middlewares.EarlyExitOnPreflightRequests())
	r.Use(middlewares.SetCorsPolicy(config.AllowedOrigins))
	r.Use(middlewares.SetRequestValues(*db, config))
	handlers.SetupRoutes(r, fr)

	pprof.Register(r)

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}

func seedData(db *gorm.DB) {
	sqldb, err := db.DB()
	if err != nil {
		panic(err)
	}

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

	user := models.User{}
	if err := models.ParseModelWithUnmarshal(&user, "user.json"); err != nil {
		panic(err)
	}
	db.Create(&user)

	rooms := make([]models.Room, 0, 3)
	if err := models.ParseModelWithUnmarshal(&rooms, "rooms.json"); err != nil {
	}
	db.Create(&rooms)
}

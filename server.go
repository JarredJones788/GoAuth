package main

import (
	"auth"
	"db"
	"fmt"
	"os"
	"router"
	"types"
)

func getConfig(mode string) *types.Config {
	if mode == "development" {
		return &types.Config{
			MySQL: types.MySQLConfig{
				Conn:     "localhost",
				UserName: "root",
				Password: "",
				DBName:   "boxing",
			},
			Redis: types.RedisConfig{
				Conn:    "localhost:6379",
				Timeout: 1800, //How long cache key lasts (Seconds)
				Active:  false,
			},
			Email: types.EmailConfig{
				Email:       "info@tester788.info",
				Password:    "n!_e6b62ft$3fd",
				SMTPAddress: "smtp.office365.com",
				SMTPPort:    587,
			},
			ServerPort:  ":4000",
			Host:        "http://localhost:3000",
			LogDuration: 30, //Days
		}
	}
	return &types.Config{
		MySQL: types.MySQLConfig{
			Conn:     "localhost",
			UserName: "root",
			Password: "",
			DBName:   "boxing",
		},
		Redis: types.RedisConfig{
			Conn:    "localhost:6379",
			Timeout: 1800, //How long cache key lasts (Seconds)
			Active:  false,
		},
		Email: types.EmailConfig{
			Email:       "info@tester788.info",
			Password:    "n!_e6b62ft$3fd",
			SMTPAddress: "smtp.office365.com",
			SMTPPort:    587,
		},
		ServerPort:  ":4000",
		Host:        "http://localhost:3000",
		LogDuration: 30, //Days
	}
}

func main() {

	//Default Dev config
	configType := "development"

	//Check if production flag is passed.
	if len(os.Args) > 1 {
		if os.Args[1] == "-production" {
			configType = "production"
		}
	}

	//Get correct config.
	config := getConfig(configType)

	db := db.MySQL{}.Init(config)
	if db == nil {
		fmt.Println("Cannot connect to database")
		return
	}
	auth := auth.Authenticate{}.Init(db, config)
	router.Router{}.Init(auth, config)
}

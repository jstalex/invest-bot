package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type RobotConfig struct {
	Token    string
	AppName  string
	EndPoint string
}

func NewRobotConfig(filename string) *RobotConfig {
	var c RobotConfig
	if err := godotenv.Load(filename); err != nil {
		log.Print("No .env file found")
	}
	c.Token = stringFromEnv("TINKOFF_TOKEN")
	c.EndPoint = stringFromEnv("END_POINT")
	c.AppName = stringFromEnv("APP_NAME")
	return &c
}
func stringFromEnv(key string) string {
	answer, err := os.LookupEnv(key)
	if !err {
		log.Print("robot config reading error")
	}
	return answer
}

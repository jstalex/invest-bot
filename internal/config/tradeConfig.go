package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type tradeConfig struct {
	Tradeinstrument string `yaml:"figi"`
	AccountID       string `yaml:"id"`
}

func NewTradeConfig(filename string) *tradeConfig {
	var t tradeConfig
	input, err := os.ReadFile(filename)
	if err != nil {
		log.Println(err)
	}
	err = yaml.Unmarshal(input, &t)
	if err != nil {
		log.Println(err)
	}
	return &t
}

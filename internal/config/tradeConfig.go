package config

import (
	"gopkg.in/yaml.v3"
	//s "invest-bot/internal/sdk"
	"log"
	"os"
)

type TradeConfig struct {
	TradeInstruments []string `yaml:"instruments"`
	AccountID        string   `yaml:"id"`
	Strategy         string   `yaml:"strategy"`
	Period           int      `yaml:"period"`
	Window           int      `yaml:"window"`
}

func LoadTradeConfig(filename string) *TradeConfig {
	var t TradeConfig
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

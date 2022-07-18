package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	pb "invest-bot/api/proto"
	s "invest-bot/internal/sdk"
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
func (tc *TradeConfig) CreateAccountID(sdk *s.SDK) {
	if tc.AccountID == "" {
		fmt.Println("Sandbox account id field is empty in tradeconfig")
		openAccountReq, err := sdk.Sandbox.OpenSandboxAccount(sdk.Ctx, &pb.OpenSandboxAccountRequest{})
		if err != nil {
			log.Println("sandbox account opening error", err)
		}
		accountID := openAccountReq.AccountId
		tc.AccountID = accountID
		sdk.TradeConfig = tc
		cnf, err := yaml.Marshal(tc)
		if err != nil {
			log.Println(err)
		}
		err = os.WriteFile("config.yaml", cnf, 0666)
		if err != nil {
			log.Println(err)
		}
		fmt.Println("Account ID was successfully created")
	}
}

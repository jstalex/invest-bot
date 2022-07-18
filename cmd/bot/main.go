package main

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	investapi "invest-bot/api/proto"
	"invest-bot/internal/config"
	"invest-bot/internal/engine"
	s "invest-bot/internal/sdk"
	"invest-bot/internal/trader"
	"log"
	"os"
)

func main() {
	rc := config.LoadRobotConfig("config.env")
	tc := config.LoadTradeConfig("config.yaml")
	sdk := s.NewSDK(rc, tc)
	if tc.AccountID == "" {
		fmt.Println("Sandbox account id field is empty in tradeconfig")
		openAccountReq, err := sdk.Sandbox.OpenSandboxAccount(context.Background(), &investapi.OpenSandboxAccountRequest{})
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
	}
	traders := trader.LoadTradersFromConfig(sdk, tc)
	engine.RunOnSandbox(sdk, traders)
	//engine.TestOnHisoryData(sdk, traders, "2022-07-14", "2022-07-15")
}

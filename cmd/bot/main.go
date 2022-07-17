package main

import (
	"invest-bot/internal/config"
	s "invest-bot/internal/sdk"
	"invest-bot/internal/trader"
)

type OrderSide int

const (
	BUY OrderSide = iota
	SELL
)

func main() {
	rc := config.LoadRobotConfig("config.env")
	tc := config.LoadTradeConfig("config.yml")
	sdk := s.NewSDK(rc, tc)
	traders := trader.LoadTradersFromConfig(sdk, tc)

	traders[""].RunOnSandbox(sdk)
	/*for _, t := range traders {
		t.TestOnHisoricalData(sdk, "2022-07-13", "2022-07-14")
	}*/
}

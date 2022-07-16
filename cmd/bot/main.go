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
	rc := config.NewRobotConfig("config.env")
	tc := config.LoadTradeConfig("config.yml")
	sdk := s.NewSDK(rc)
	traders := trader.InitTradersFromConfig(tc)
	for _, t := range traders {
		t.TestOnHisoricalData(sdk, "2021-10-12", "2021-10-13")
	}
}

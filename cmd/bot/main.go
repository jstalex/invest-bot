package main

import (
	"invest-bot/internal/config"
	"invest-bot/internal/engine"
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
	engine.RunOnSandbox(sdk, traders)
	//engine.TestOnHisoryData(sdk, traders, "2022-07-14", "2022-07-15")
}

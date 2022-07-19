package main

import (
	"invest-bot/internal/config"
	"invest-bot/internal/engine"
	s "invest-bot/internal/sdk"
	"invest-bot/internal/trader"
)

func main() {
	rc := config.LoadRobotConfig("config.env")
	tc := config.LoadTradeConfig("config.yaml")
	sdk := s.NewSDK(rc, tc)
	// если аккаунта в файле конфига нет, то он создастся и запишется в файл
	sdk.CreateSandboxAccount(tc)
	traders := trader.LoadTradersFromConfig(sdk, tc)
	engine.RunOnSandbox(sdk, traders)
	//engine.TestOnHisoryData(sdk, traders, "2022-07-14", "2022-07-15")
}

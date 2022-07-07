package main

import (
	"fmt"
	"invest-bot/internal/config"
)

func main() {
	rc := config.NewRobotConfig("config.env")
	//tc := config.NewTradeConfig("config.yml")
	//sdk := s.NewSDK(rc)
	//sdk.Sandbox.GetSandboxPortfolio(sdk.Ctx)
	fmt.Println(rc)
}

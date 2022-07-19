package engine

import (
	"fmt"
	s "invest-bot/internal/sdk"
	"invest-bot/internal/trader"
	"log"
)

func SelectMode(sdk *s.SDK, traders map[string]*trader.Trader) {
	fmt.Print("Modes:\n 1. Run on history data \n 2. Run on sandbox \nEnter mod: ")
	var mode int
	_, err := fmt.Scan(&mode)
	if err != nil {
		log.Println("mode scanning error", err)
	}
	switch mode {
	case 1:
		fmt.Print("Enter start, stop day in format YYYY-MM-DD:")
		var start, stop string
		_, err := fmt.Scan(&start, &stop)
		if err != nil {
			log.Println("date scanning error", err)
		}
		TestOnHisoryData(sdk, traders, start, stop)
	case 2:
		RunOnSandbox(sdk, traders)
	}
}

package engine

import (
	s "invest-bot/internal/sdk"
	"invest-bot/internal/trader"
)

func TestOnHisoryData(sdk *s.SDK, traders map[string]*trader.Trader, from string, to string) {
	for _, t := range traders {
		t.TestOnHisoricalData(sdk, from, to)
	}
}

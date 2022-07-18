package engine

import (
	s "invest-bot/internal/sdk"
	"invest-bot/internal/trader"
)

func RunOnSandbox(sdk *s.SDK, subscribers map[string]*trader.Trader) {
	CandlesFromStream(sdk, subscribers)
}

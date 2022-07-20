package trader

import (
	"fmt"
	s "invest-bot/internal/sdk"
)

func (t *Trader) TestOnHisoricalData(sdk *s.SDK, from string, to string) {
	dataset := sdk.GetHistoricalCandles(t.Figi, t.Period, from, to)
	for _, hc := range dataset {
		if t.series.AddCandle(sdk.HistoricalCandleToTechan(hc)) {
			if t.ruleStrategy.ShouldEnter(t.series.LastIndex(), t.Record) {
				t.AddTrade(BUY, t.series.LastCandle().ClosePrice.Float(), t.series.LastCandle().ClosePrice.Float(), t.series.LastCandle().Period.End)
			} else if t.ruleStrategy.ShouldExit(t.series.LastIndex(), t.Record) {
				t.AddTrade(SELL, t.series.LastCandle().ClosePrice.Float(), t.series.LastCandle().ClosePrice.Float(), t.series.LastCandle().Period.End)
			} // else hold instrument
		} else {
			fmt.Println("candle adding error")
		}
	}
	profit := 0.0
	for _, trade := range t.Record.Trades {
		res := trade.ExitOrder().Price.Sub(trade.EntranceOrder().Price).Float()
		profit += res
	}
	fmt.Println("total profit with ", t.Figi, ": ", profit)
}

package trader

import (
	"fmt"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"invest-bot/internal/algo"
	"invest-bot/internal/config"
	s "invest-bot/internal/sdk"
)

type Trader struct {
	Figi     string
	Strategy string
	Period   int
	Window   int
}

func NewTrader(i int, cnf *config.TradeConfig) *Trader {
	return &Trader{
		Figi:     cnf.TradeInstruments[i],
		Strategy: cnf.Strategy,
		Period:   cnf.Period,
		Window:   cnf.Window,
	}
}
func InitTradersFromConfig(cnf *config.TradeConfig) []*Trader {
	traders := make([]*Trader, 0, 0)
	for i, _ := range cnf.TradeInstruments {
		traders = append(traders, NewTrader(i, cnf))
	}
	return traders
}
func (t *Trader) TestOnHisoricalData(sdk *s.SDK, from string, to string) {
	var ruleStrategy *techan.RuleStrategy
	var series *techan.TimeSeries
	var record *techan.TradingRecord

	switch t.Strategy {
	case "aroon":
		ruleStrategy, series, record = algo.NewAroonStrategy(t.Window)
	case "ema":
		ruleStrategy, series, record = algo.NewEMAStrategy(t.Window)
	case "classic":
		ruleStrategy, series, record = algo.NewClassicStrategy(t.Window)
	default:
		ruleStrategy, series, record = algo.NewEMAStrategy(t.Window)
	}

	dataset := sdk.GetHistoricalCandles(t.Figi, t.Period, from, to)
	for _, hc := range dataset {
		if series.AddCandle(sdk.HistoricalCandleToTechan(hc)) {
			if ruleStrategy.ShouldEnter(series.LastIndex(), record) {
				record.Operate(techan.Order{
					Side:          techan.BUY,
					Security:      "testid",
					Price:         series.LastCandle().ClosePrice,
					Amount:        big.NewDecimal(5),
					ExecutionTime: series.LastCandle().Period.End,
				})
			} else if ruleStrategy.ShouldExit(series.LastIndex(), record) {
				record.Operate(techan.Order{
					Side:          techan.SELL,
					Security:      "testid",
					Price:         series.LastCandle().ClosePrice,
					Amount:        record.CurrentPosition().ExitValue(),
					ExecutionTime: series.LastCandle().Period.End,
				})
			} // else hold instrument
		} else {
			fmt.Println("candle adding error")
		}
	}

	income := 0.0
	for _, trade := range record.Trades {
		res := trade.ExitOrder().Price.Sub(trade.EntranceOrder().Price).Float()
		//fmt.Printf("Order number %v %v\n", i+1, res)
		income += res
	}
	fmt.Println("total profit with ", t.Figi, ": ", income)
}

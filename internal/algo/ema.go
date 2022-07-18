package algo

import "github.com/sdcoffey/techan"

func NewEMAStrategy(w int) (*techan.RuleStrategy, *techan.TimeSeries, *techan.TradingRecord) {
	series := techan.NewTimeSeries()
	record := techan.NewTradingRecord()

	closePrices := techan.NewClosePriceIndicator(series)
	movingAverage := techan.NewEMAIndicator(closePrices, w)

	entryRule := techan.And( // правило входа
		techan.NewCrossUpIndicatorRule(movingAverage, closePrices),
		techan.PositionNewRule{})
	exitRule := techan.And( // правило выхода
		techan.NewCrossDownIndicatorRule(closePrices, movingAverage),
		techan.PositionOpenRule{})
	ruleStrategy := &techan.RuleStrategy{
		UnstablePeriod: w,
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}
	return ruleStrategy, series, record
}

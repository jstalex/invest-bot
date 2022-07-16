package algo

import (
	"github.com/sdcoffey/techan"
)

func NewAroonStrategy(w int) (*techan.RuleStrategy, *techan.TimeSeries, *techan.TradingRecord) {
	series := techan.NewTimeSeries()
	record := techan.NewTradingRecord()

	lowPrices := techan.NewLowPriceIndicator(series)
	highPrices := techan.NewHighPriceIndicator(series)

	aroonDownIndicator := techan.NewAroonDownIndicator(lowPrices, w)
	aroonUpIndicator := techan.NewAroonUpIndicator(highPrices, w)

	entryRule := techan.And(
		techan.NewCrossUpIndicatorRule(aroonDownIndicator, aroonUpIndicator),
		techan.PositionNewRule{})
	exitRule := techan.And(
		techan.NewCrossDownIndicatorRule(aroonUpIndicator, aroonDownIndicator),
		techan.PositionOpenRule{})

	ruleStrategy := &techan.RuleStrategy{
		UnstablePeriod: w,
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}
	return ruleStrategy, series, record
}

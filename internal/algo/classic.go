package algo

import "github.com/sdcoffey/techan"

func NewClassicStrategy(w int) (*techan.RuleStrategy, *techan.TimeSeries, *techan.TradingRecord) {
	series := techan.NewTimeSeries()
	record := techan.NewTradingRecord()

	lowPrices := techan.NewLowPriceIndicator(series)
	highPrices := techan.NewHighPriceIndicator(series)

	closePrices := techan.NewClosePriceIndicator(series)
	movingAverage := techan.NewEMAIndicator(closePrices, w)

	aroonDownIndicator := techan.NewAroonDownIndicator(lowPrices, w)
	aroonUpIndicator := techan.NewAroonUpIndicator(highPrices, w)

	customEntry := techan.And(techan.NewCrossUpIndicatorRule(aroonDownIndicator, aroonUpIndicator),
		techan.NewCrossUpIndicatorRule(movingAverage, closePrices))

	customExit := techan.And(techan.NewCrossDownIndicatorRule(aroonUpIndicator, aroonDownIndicator),
		techan.NewCrossDownIndicatorRule(closePrices, movingAverage))

	entryRule := techan.And(
		customEntry,
		techan.PositionNewRule{})
	exitRule := techan.And(
		customExit,
		techan.PositionOpenRule{})

	ruleStrategy := &techan.RuleStrategy{
		UnstablePeriod: w,
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}
	return ruleStrategy, series, record
}

package algo

import "github.com/sdcoffey/techan"

func NewClassicStrategy(w int) (*techan.RuleStrategy, *techan.TimeSeries, *techan.TradingRecord) {
	series := techan.NewTimeSeries()
	record := techan.NewTradingRecord()

	indicator := techan.NewClosePriceIndicator(series)
	entryConstant := techan.NewConstantIndicator(21600)
	exitConstant := techan.NewConstantIndicator(21400)

	// Is satisfied when the price ema moves above 30 and the current position is new
	entryRule := techan.And(
		techan.NewCrossUpIndicatorRule(entryConstant, indicator),
		techan.PositionNewRule{})

	// Is satisfied when the price ema moves below 10 and the current position is open
	exitRule := techan.And(
		techan.NewCrossDownIndicatorRule(indicator, exitConstant),
		techan.PositionOpenRule{})

	ruleStrategy := &techan.RuleStrategy{
		UnstablePeriod: w, // Period before which ShouldEnter and ShouldExit will always return false
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}
	return ruleStrategy, series, record
}

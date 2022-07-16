package algo

import "github.com/sdcoffey/techan"

func NewEMAStrategy(w int) (*techan.RuleStrategy, *techan.TimeSeries, *techan.TradingRecord) {
	series := techan.NewTimeSeries()
	record := techan.NewTradingRecord()

	closePrices := techan.NewClosePriceIndicator(series)    // отсеивает High, Low, Open, на выходе только Close
	movingAverage := techan.NewEMAIndicator(closePrices, w) // Создает экспоненциальное среднее с окном в n свечей

	entryRule := techan.And( // правило входа
		techan.NewCrossUpIndicatorRule(movingAverage, closePrices), // когда свеча закрытия пересечет EMA (станет выше EMA)
		techan.PositionNewRule{})                                   // и сделок не открыто — мы покупаем
	exitRule := techan.And( // правило выхода
		techan.NewCrossDownIndicatorRule(closePrices, movingAverage), // когда свеча закроется ниже EMA
		techan.PositionOpenRule{})                                    // и сделка открыта — продаем
	ruleStrategy := &techan.RuleStrategy{
		UnstablePeriod: w,
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}
	return ruleStrategy, series, record
}

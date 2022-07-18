package sdk

import (
	pb "invest-bot/api/proto"
	"log"
)

func (s *SDK) GetHistoricalCandles(figi string, period int, from string, to string) []*pb.HistoricCandle {
	response, err := s.Marketdata.GetCandles(s.Ctx, &pb.GetCandlesRequest{
		Figi:     figi,
		From:     s.DateToTimestamp(from),
		To:       s.DateToTimestamp(to),
		Interval: pb.CandleInterval(period),
	})
	if err != nil {
		log.Println("candles not found ", err)
	}
	return response.GetCandles()
}
func (s *SDK) PostSandboxOrder(figi string, direction pb.OrderDirection) (float64, bool) {
	resp, err := s.Sandbox.PostSandboxOrder(s.Ctx, &pb.PostOrderRequest{
		Figi:      figi,
		Quantity:  1,
		Price:     nil,
		Direction: direction,
		AccountId: s.TradeConfig.AccountID,
		OrderType: pb.OrderType_ORDER_TYPE_MARKET,
		OrderId:   "testid",
	})
	if err != nil {
		log.Println("sandbox post order error:", err)
	}
	ok := resp.ExecutionReportStatus == pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL
	return s.MoneyValueToFloat(resp.TotalOrderAmount), ok
}

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

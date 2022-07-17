package sdk

import (
	pb "invest-bot/api/proto"
	"invest-bot/internal/trader"
	"log"
	"time"
)

func (s *SDK) CandlesFromStream(subscribers map[string]*trader.Trader) *pb.Candle {
	instruments := make([]*pb.CandleInstrument, 0, 0)
	for _, t := range s.TradeConfig.TradeInstruments {
		instruments = append(instruments, &pb.CandleInstrument{
			Figi:     t,
			Interval: pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_MINUTE,
		})
	}
	for {
		err := s.MarketStream.Send(&pb.MarketDataRequest{
			Payload: &pb.MarketDataRequest_SubscribeCandlesRequest{
				SubscribeCandlesRequest: &pb.SubscribeCandlesRequest{
					SubscriptionAction: pb.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
					Instruments:        instruments,
					WaitingClose:       false,
				}}})
		if err != nil {
			log.Println(err)
		}
		resp, err := s.MarketStream.Recv()
		if err != nil {
			log.Println(err)
		}
		// bcs period = 1 min
		time.Sleep(time.Minute)
		if _, ok := subscribers[resp.GetCandle().GetFigi()]; ok {
			subscribers[resp.GetCandle().GetFigi()].HandleIncomingCandle(resp.GetCandle())
		} else {
			log.Println("not found subscriber to instrument ", resp.GetCandle().GetFigi())
		}

	}
}

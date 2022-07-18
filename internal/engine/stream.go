package engine

import (
	pb "invest-bot/api/proto"
	s "invest-bot/internal/sdk"
	"invest-bot/internal/trader"
	"log"
	"time"
)

func CandlesFromStream(sdk *s.SDK, subscribers map[string]*trader.Trader) {
	streamClient := pb.NewMarketDataStreamServiceClient(sdk.Conn)
	marketStream, err := streamClient.MarketDataStream(sdk.Ctx)
	if err != nil {
		log.Println("marketdata stream error", err)
	}
	// создание слайса инструментов для зпроса о подписке
	instruments := make([]*pb.CandleInstrument, 0, 0)
	for _, t := range sdk.TradeConfig.TradeInstruments {
		instruments = append(instruments, &pb.CandleInstrument{
			Figi:     t,
			Interval: pb.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_MINUTE,
		})
	}
	// запрос на подписку
	err = marketStream.Send(&pb.MarketDataRequest{
		Payload: &pb.MarketDataRequest_SubscribeCandlesRequest{
			SubscribeCandlesRequest: &pb.SubscribeCandlesRequest{
				SubscriptionAction: pb.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
				Instruments:        instruments,
				WaitingClose:       true,
			}}})
	if err != nil {
		log.Println(err)
	}
	// слушаем стрим и вызываем обработку свечи у соответсвующего иструменту трейдера
	for {
		resp, err := marketStream.Recv()
		if err != nil {
			log.Println(err)
		}
		if resp.GetCandle() != nil {
			if _, ok := subscribers[resp.GetCandle().GetFigi()]; ok {
				go subscribers[resp.GetCandle().GetFigi()].HandleIncomingCandle(resp.GetCandle())
			} else {
				log.Println("not found subscriber to instrument ", resp.GetCandle().Figi)
			}
		} else {
			log.Println("response do not contain a candle")
		}
		time.Sleep(time.Second)
	}
}

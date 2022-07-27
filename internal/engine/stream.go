package engine

import (
	pb "invest-bot/api/proto"
	s "invest-bot/internal/sdk"
	"invest-bot/internal/trader"
	"log"
	"sync"
	"time"
)

func ListenCandlesFromStream(sdk *s.SDK, subscribers map[string]*trader.Trader) {
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
	// слушаем стрим определенное время и вызываем обработку свечи у, соответсвующего инструменту, трейдера
	stopTime := time.Now().Add(time.Duration(sdk.TradeConfig.TradingDuration) * time.Minute)
	// функция дождется завершения обработки всех свечей в горутинах
	var wg sync.WaitGroup
	var incomingCandle *pb.Candle
	for time.Now().Before(stopTime) {
		resp, err := marketStream.Recv()
		if err != nil {
			log.Fatalln(err)
		}
		incomingCandle = resp.GetCandle()
		if incomingCandle != nil {
			if _, ok := subscribers[incomingCandle.GetFigi()]; ok {
				// вызываем обработку свечи у трейдера
				wg.Add(1)
				go func() {
					subscribers[incomingCandle.GetFigi()].HandleIncomingCandle(incomingCandle)
					wg.Done()
				}()
			} else {
				log.Println("not found subscriber to instrument ", incomingCandle.GetFigi())
			}
		} else {
			log.Println("response do not contain a candle")
		}
	}
	wg.Wait()
}

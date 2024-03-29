package engine

import (
	"fmt"
	"github.com/golang/protobuf/ptypes/timestamp"
	pb "invest-bot/api/proto"
	s "invest-bot/internal/sdk"
	"invest-bot/internal/trader"
	"log"
	"time"
)

// Run - Запуск робота
func Run(sdk *s.SDK, subscribers map[string]*trader.Trader) {
	checkSchedule(sdk)
	initailBalance := sdk.GetMoneyBalance()
	// подписаться на свечи инструментов
	ListenCandlesFromStream(sdk, subscribers)
	// в конце торговли закрываем все активные позиции
	finishTradingSession(sdk, subscribers)
	// вычисляем разницу между начальным балансом на счете и итоговым
	balanceAfterTrading := sdk.GetMoneyBalance()
	fmt.Println("Report:")
	// отчет о работе
	var techanTotalProfit float64 = 0
	for _, t := range subscribers {
		positions := t.Record.Trades
		fmt.Println("Instrument: ", sdk.GetTickerByFigi(t.Figi))
		for i, p := range positions {
			fmt.Print("position number ", i, ", ")
			fmt.Println("Profit:", p.ExitOrder().Amount.Sub(p.EntranceOrder().Amount))
			techanTotalProfit += p.ExitOrder().Amount.Float() - p.EntranceOrder().Amount.Float()
		}
		fmt.Print("\n")
	}
	fmt.Println("Profit after trading session, by balance =", balanceAfterTrading-initailBalance, "RUB")
	fmt.Println("Profit after trading session, by trading record =", techanTotalProfit)
}

// закрытие всех текущих позиций на счете
func finishTradingSession(sdk *s.SDK, subscribers map[string]*trader.Trader) {
	var positionsResp *pb.PositionsResponse
	var err error
	if sdk.TradingMode == s.Sandbox {
		positionsResp, err = sdk.Sandbox.GetSandboxPositions(sdk.Ctx, &pb.PositionsRequest{AccountId: sdk.TradeConfig.AccountID})
		if err != nil {
			log.Println(err)
		}
	} else if sdk.TradingMode == s.Real {
		positionsResp, err = sdk.Operations.GetPositions(sdk.Ctx, &pb.PositionsRequest{AccountId: sdk.TradeConfig.RealAccountID})
		if err != nil {
			log.Println(err)
		}
	}
	for _, p := range positionsResp.Securities {
		subscribers[p.Figi].Sell(p.Balance / sdk.GetLotsByFigi(p.GetFigi()))
	}
	log.Println("Trading session finished")
}

func checkSchedule(sdk *s.SDK) {
	tradingscheduleResp, err := sdk.Instruments.TradingSchedules(sdk.Ctx, &pb.TradingSchedulesRequest{
		Exchange: "MOEX",
		From:     &timestamp.Timestamp{Seconds: time.Now().UTC().Unix(), Nanos: 0},
		To:       &timestamp.Timestamp{Seconds: time.Now().Add(time.Hour * 8).UTC().Unix(), Nanos: 0},
	})
	if err != nil {
		log.Println("schedule checking error:", err)
	}
	var tradingDays []*pb.TradingDay
	for _, sched := range tradingscheduleResp.GetExchanges() {
		if sched.GetExchange() == "MOEX" {
			tradingDays = sched.GetDays()
			break
		}
	}
	for _, day := range tradingDays {
		if !day.IsTradingDay || time.Now().After(day.EndTime.AsTime()) {
			log.Fatalln("Trading isn't available, exchange is closed")
		}
	}
}

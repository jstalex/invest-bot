package engine

import (
	"fmt"
	pb "invest-bot/api/proto"
	s "invest-bot/internal/sdk"
	"invest-bot/internal/trader"
	"log"
)

// RunOnSandbox - Запуск робота в песочнице
func RunOnSandbox(sdk *s.SDK, subscribers map[string]*trader.Trader) {
	initailBalance := sdk.GetSandboxMoneyBalance()
	// подписаться на свечи инструментов
	CandlesFromStream(sdk, subscribers)
	// в конце торговли закрываем все активные позиции
	allPositionsAreClosed := finishTradingSession(sdk)
	if allPositionsAreClosed {
		log.Println("Trading session successful finished")
	} else {
		log.Println("error in final selling of instruments")
	}
	// вычисляем разницу между начальным балансом на счете и итоговым
	balanceAfterTrading := sdk.GetSandboxMoneyBalance()
	fmt.Println("Profit after trading session =", balanceAfterTrading-initailBalance, "RUB")
}

// закрытие всех текущих позиций на счете
func finishTradingSession(sdk *s.SDK) bool {
	positionsResp, err := sdk.Sandbox.GetSandboxPositions(sdk.Ctx, &pb.PositionsRequest{AccountId: sdk.TradeConfig.AccountID})
	if err != nil {
		log.Println(err)
	}
	for _, p := range positionsResp.Securities {
		_, ok := sdk.PostSandboxOrder(p.Figi, p.Balance/sdk.GetLotsByFigi(p.GetFigi()), pb.OrderDirection_ORDER_DIRECTION_SELL)
		if !ok {
			return false
		}
	}
	return true
}

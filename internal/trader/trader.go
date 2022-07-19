package trader

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	pb "invest-bot/api/proto"
	"invest-bot/internal/algo"
	"invest-bot/internal/config"
	s "invest-bot/internal/sdk"
	"log"
	"time"
)

type OrderSide int

const (
	BUY OrderSide = iota
	SELL
	HOLD
)

type Trader struct {
	Figi   string
	Period int

	ruleStrategy *techan.RuleStrategy
	series       *techan.TimeSeries
	record       *techan.TradingRecord

	sdk *s.SDK
}

func NewTrader(i int, cnf *config.TradeConfig, s *s.SDK) *Trader {

	var ruleStrategy *techan.RuleStrategy
	var series *techan.TimeSeries
	var record *techan.TradingRecord

	switch cnf.Strategy {
	case "aroon":
		ruleStrategy, series, record = algo.NewAroonStrategy(cnf.Window)
	case "ema":
		ruleStrategy, series, record = algo.NewEMAStrategy(cnf.Window)
	case "classic":
		ruleStrategy, series, record = algo.NewClassicStrategy(cnf.Window)
	default:
		ruleStrategy, series, record = algo.NewEMAStrategy(cnf.Window)
	}

	return &Trader{
		Figi:         cnf.TradeInstruments[i],
		Period:       cnf.Period,
		ruleStrategy: ruleStrategy,
		series:       series,
		record:       record,
		sdk:          s,
	}
}

// LoadTradersFromConfig - создание экземпляров трейдеров для каждого инструмента из конфига
func LoadTradersFromConfig(s *s.SDK, cnf *config.TradeConfig) map[string]*Trader {
	traders := make(map[string]*Trader)
	for i, f := range cnf.TradeInstruments {
		traders[f] = NewTrader(i, cnf, s)
	}
	return traders
}

// HandleIncomingCandle - обработка свечи из Marketdata stream
func (t *Trader) HandleIncomingCandle(inputCandle *pb.Candle) {
	// если пришедшая свеча удочно добавлена, то принимаем дальнейшее решение
	if t.series.AddCandle(t.sdk.PBCandleToTechan(inputCandle)) {
		if t.selectOperation() == BUY && t.possibleToBuy(t.sdk.QuotationToFloat(inputCandle.High), t.sdk.TradeConfig.LotsQuantity) {
			executedPrice, ok := t.sdk.PostSandboxOrder(t.Figi, t.sdk.TradeConfig.LotsQuantity, pb.OrderDirection_ORDER_DIRECTION_BUY)
			if ok {
				t.addTrade(BUY, executedPrice, executedPrice, time.Now())
				log.Println("Buy order executed with figi ", t.Figi)
			} else {
				log.Println("buy order error")
			}
		} else if t.selectOperation() == SELL && t.possibleToSell(t.sdk.TradeConfig.LotsQuantity) {
			executedPrice, ok := t.sdk.PostSandboxOrder(t.Figi, t.sdk.TradeConfig.LotsQuantity, pb.OrderDirection_ORDER_DIRECTION_SELL)
			if ok {
				t.addTrade(SELL, executedPrice, executedPrice, time.Now())
				log.Println("sell order executed with figi ", t.Figi)
			} else {
				log.Println("sell order error")
			}
		}
	} else {
		log.Println("candle adding error")
	}
}

// принятие решения о покупке/продаже при помощи торговой стратегии
func (t *Trader) selectOperation() OrderSide {
	if t.ruleStrategy.ShouldEnter(t.series.LastIndex(), t.record) {
		return BUY
	} else if t.ruleStrategy.ShouldExit(t.series.LastIndex(), t.record) {
		return SELL
	} else {
		return HOLD
	}
}

// проврека возможности покупки инструмента на счет песочницы
func (t *Trader) possibleToBuy(price float64, quantity int64) bool {
	var totalSum float64
	shareResp, err := t.sdk.Instruments.ShareBy(t.sdk.Ctx, &pb.InstrumentRequest{
		IdType: pb.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI,
		Id:     t.Figi,
	})
	if err != nil {
		log.Println(err)
	}
	portfolioResp, err := t.sdk.Sandbox.GetSandboxPortfolio(t.sdk.Ctx, &pb.PortfolioRequest{AccountId: t.sdk.TradeConfig.AccountID})
	if err != nil {
		log.Println(err)
	}
	// цена 1 инструмента * количество инструментов в лоте * кол-во лотов
	totalSum = price * float64(quantity) * float64(shareResp.GetInstrument().Lot)
	// если иструмент доступен для торговли на бирже и хватает денег на его покупку
	return shareResp.GetInstrument().ApiTradeAvailableFlag && t.sdk.MoneyValueToFloat(portfolioResp.GetTotalAmountCurrencies()) >= totalSum
}

// проверка аозможности продажи инструмента со счета песочницы
func (t *Trader) possibleToSell(quantity int64) bool {
	shareResp, err := t.sdk.Instruments.ShareBy(t.sdk.Ctx, &pb.InstrumentRequest{
		IdType: pb.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI,
		Id:     t.Figi,
	})
	if err != nil {
		log.Println(err)
	}
	positionsResp, err := t.sdk.Sandbox.GetSandboxPositions(t.sdk.Ctx, &pb.PositionsRequest{AccountId: t.sdk.TradeConfig.AccountID})
	if err != nil {
		log.Println(err)
	}
	contain := false
	var LotsOnBalance int64
	for _, p := range positionsResp.Securities {
		if p.Figi == t.Figi {
			contain = true
			LotsOnBalance = p.Balance / int64(shareResp.GetInstrument().Lot)
			break
		}
	}
	// если инструмент доступен для торговли на бирже, есть ли он у нас и достаточно ли для продажи
	return shareResp.GetInstrument().ApiTradeAvailableFlag && contain && LotsOnBalance >= quantity
}

// добавляем исполненные поручения в запись
func (t *Trader) addTrade(side OrderSide, price float64, amount float64, time time.Time) {
	if side == BUY {
		t.record.Operate(techan.Order{
			Side:          techan.BUY,
			Security:      "testid",
			Price:         big.NewDecimal(price),
			Amount:        big.NewDecimal(amount),
			ExecutionTime: time,
		})
	} else if side == SELL {
		t.record.Operate(techan.Order{
			Side:          techan.SELL,
			Security:      "testid",
			Price:         big.NewDecimal(price),
			Amount:        big.NewDecimal(amount),
			ExecutionTime: time,
		})
	}
}

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
	Figi         string
	Period       int
	Record       *techan.TradingRecord
	ruleStrategy *techan.RuleStrategy
	series       *techan.TimeSeries
	sdk          *s.SDK
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
		Record:       record,
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
	// если пришедшая свеча удачно добавлена, то принимаем дальнейшее решение
	if t.series.AddCandle(t.sdk.PBCandleToTechan(inputCandle)) {
		if t.selectOperation() == BUY && t.possibleToBuy(t.sdk.QuotationToFloat(inputCandle.High)) {
			t.Buy(t.sdk.TradeConfig.LotsQuantity)
		} else if t.selectOperation() == SELL && t.possibleToSell() {
			t.Sell(t.sdk.TradeConfig.LotsQuantity)
		}
	} else {
		log.Println("candle adding error")
	}
}

// принятие решения о покупке/продаже при помощи торговой стратегии
func (t *Trader) selectOperation() OrderSide {
	if t.ruleStrategy.ShouldEnter(t.series.LastIndex(), t.Record) {
		return BUY
	} else if t.ruleStrategy.ShouldExit(t.series.LastIndex(), t.Record) {
		return SELL
	} else {
		return HOLD
	}
}

// проврека возможности покупки инструмента на счет песочницы/реальный
func (t *Trader) possibleToBuy(price float64) bool {
	var totalSum float64
	instrumentResp, err := t.sdk.Instruments.GetInstrumentBy(t.sdk.Ctx, &pb.InstrumentRequest{
		IdType: pb.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI,
		Id:     t.Figi,
	})
	if err != nil {
		log.Println(err)
	}
	var portfolioResp *pb.PortfolioResponse
	if t.sdk.TradingMode == s.Sandbox {
		portfolioResp, err = t.sdk.Sandbox.GetSandboxPortfolio(t.sdk.Ctx, &pb.PortfolioRequest{AccountId: t.sdk.TradeConfig.AccountID})
		if err != nil {
			log.Println(err)
		}
	} else if t.sdk.TradingMode == s.Real {
		portfolioResp, err = t.sdk.Operations.GetPortfolio(t.sdk.Ctx, &pb.PortfolioRequest{AccountId: t.sdk.TradeConfig.RealAccountID})
		if err != nil {
			log.Println(err)
		}
	}
	// сумма = цена 1 инструмента * количество инструментов в лоте * кол-во лотов
	totalSum = price * float64(t.sdk.TradeConfig.LotsQuantity) * float64(instrumentResp.GetInstrument().Lot)
	// если иструмент доступен для торговли на бирже и хватает денег на его покупку
	return instrumentResp.GetInstrument().ApiTradeAvailableFlag && t.sdk.MoneyValueToFloat(portfolioResp.GetTotalAmountCurrencies()) >= totalSum
}

// проверка возможности продажи инструмента со счета песочницы/реального
func (t *Trader) possibleToSell() bool {
	instrumentResp, err := t.sdk.Instruments.GetInstrumentBy(t.sdk.Ctx, &pb.InstrumentRequest{
		IdType: pb.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI,
		Id:     t.Figi,
	})
	if err != nil {
		log.Println(err)
	}
	var positionsResp *pb.PositionsResponse
	if t.sdk.TradingMode == s.Sandbox {
		positionsResp, err = t.sdk.Sandbox.GetSandboxPositions(t.sdk.Ctx, &pb.PositionsRequest{AccountId: t.sdk.TradeConfig.AccountID})
		if err != nil {
			log.Println(err)
		}
	} else if t.sdk.TradingMode == s.Real {
		positionsResp, err = t.sdk.Operations.GetPositions(t.sdk.Ctx, &pb.PositionsRequest{AccountId: t.sdk.TradeConfig.RealAccountID})
		if err != nil {
			log.Println(err)
		}
	}
	contain := false
	var LotsOnBalance int64
	for _, p := range positionsResp.Securities {
		if p.Figi == t.Figi {
			contain = true
			LotsOnBalance = p.Balance / int64(instrumentResp.GetInstrument().Lot)
			break
		}
	}
	// если инструмент доступен для торговли на бирже, есть ли он у нас и достаточно ли для продажи
	return instrumentResp.GetInstrument().ApiTradeAvailableFlag && contain && LotsOnBalance >= t.sdk.TradeConfig.LotsQuantity
}

// AddTrade - добавляем исполненные поручения в запись record
func (t *Trader) AddTrade(side OrderSide, price float64, amount float64, id string, time time.Time) {
	if side == BUY {
		t.Record.Operate(techan.Order{
			Side:          techan.BUY,
			Security:      id,
			Price:         big.NewDecimal(price),
			Amount:        big.NewDecimal(amount),
			ExecutionTime: time,
		})
	} else if side == SELL {
		t.Record.Operate(techan.Order{
			Side:          techan.SELL,
			Security:      id,
			Price:         big.NewDecimal(price),
			Amount:        big.NewDecimal(amount),
			ExecutionTime: time,
		})
	}
}

// Buy - метод покупки инструмента по рыночной цене, сам определяет режим работы бота
func (t *Trader) Buy(quantity int64) {
	var executedPrice, totalAmount float64
	var ok bool
	var id string
	if t.sdk.TradingMode == s.Sandbox {
		executedPrice, totalAmount, id, ok = t.sdk.PostSandboxOrder(t.Figi, quantity, pb.OrderDirection_ORDER_DIRECTION_BUY)
	} else if t.sdk.TradingMode == s.Real {
		executedPrice, totalAmount, id, ok = t.sdk.PostMarketOrder(t.Figi, quantity, pb.OrderDirection_ORDER_DIRECTION_BUY)
	}
	if ok {
		t.AddTrade(BUY, executedPrice, totalAmount, id, time.Now())
		log.Println("Buy order executed with", t.sdk.GetTickerByFigi(t.Figi), "price =", executedPrice)
	} else {
		log.Println("buy order error")
	}
}

// Sell - метод продажи инструмента по рыночной цене, сам определяет режим работы бота
func (t *Trader) Sell(quantity int64) {
	var executedPrice, totalAmount float64
	var ok bool
	var id string
	if t.sdk.TradingMode == s.Sandbox {
		executedPrice, totalAmount, id, ok = t.sdk.PostSandboxOrder(t.Figi, quantity, pb.OrderDirection_ORDER_DIRECTION_SELL)
	} else if t.sdk.TradingMode == s.Real {
		executedPrice, totalAmount, id, ok = t.sdk.PostMarketOrder(t.Figi, quantity, pb.OrderDirection_ORDER_DIRECTION_SELL)
	}
	if ok {
		t.AddTrade(SELL, executedPrice, totalAmount, id, time.Now())
		log.Println("Sell order executed with", t.sdk.GetTickerByFigi(t.Figi), "price =", executedPrice)
	} else {
		log.Println("sell order error")
	}
}

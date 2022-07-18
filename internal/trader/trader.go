package trader

import (
	"fmt"
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
func LoadTradersFromConfig(s *s.SDK, cnf *config.TradeConfig) map[string]*Trader {
	traders := make(map[string]*Trader)
	for i, f := range cnf.TradeInstruments {
		traders[f] = NewTrader(i, cnf, s)
	}
	return traders
}

func (t *Trader) HandleIncomingCandle(inputCandle *pb.Candle) {
	if t.series.AddCandle(t.sdk.PBCandleToTechan(inputCandle)) {
		if t.selectOperation() == BUY && t.possibleToBuy(t.sdk.QuotationToFloat(inputCandle.High)) {
			executedPrice, ok := t.sdk.PostSandboxOrder(t.Figi, 1)
			if ok {
				t.addTrade(BUY, executedPrice, executedPrice, time.Now())
				log.Println("Buy order executed with figi ", t.Figi)
			} else {
				log.Println("buy order error")
			}
		} else if t.selectOperation() == SELL && t.possibleToSell(1) {
			executedPrice, ok := t.sdk.PostSandboxOrder(t.Figi, 1)
			if ok {
				t.addTrade(SELL, executedPrice, executedPrice, time.Now())
				log.Println("sell order executed with figi ", t.Figi)
			} else {
				log.Println("sell order error")
			}
		}
	} else {
		fmt.Println("candle adding error")
	}
}
func (t *Trader) selectOperation() OrderSide {
	if t.ruleStrategy.ShouldEnter(t.series.LastIndex(), t.record) {
		return BUY
	} else if t.ruleStrategy.ShouldExit(t.series.LastIndex(), t.record) {
		return SELL
	} else {
		return HOLD
	}
}

func (t *Trader) possibleToBuy(sum float64) bool {
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
	return shareResp.GetInstrument().ApiTradeAvailableFlag && t.sdk.MoneyValueToFloat(portfolioResp.GetTotalAmountCurrencies()) >= sum
}

func (t *Trader) possibleToSell(amount int) bool {
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
	var balance int64
	for _, p := range positionsResp.Securities {
		if p.Figi == t.Figi {
			contain = true
			balance = p.Balance
			break
		}
	}
	return shareResp.GetInstrument().ApiTradeAvailableFlag && contain && balance >= int64(amount)
}
func (t *Trader) addTrade(side OrderSide, amount float64, price float64, time time.Time) {
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

package trader

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	pb "invest-bot/api/proto"
	"invest-bot/internal/algo"
	"invest-bot/internal/config"
	s "invest-bot/internal/sdk"
	"log"
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

func (t *Trader) RunOnSandbox(sdk *s.SDK) {

}
func (t *Trader) HandleIncomingCandle(inputCandle *pb.Candle) {
	if t.series.AddCandle(t.sdk.PBCandleToTechan(inputCandle)) {

	}
}
func (t *Trader) selectOperation() OrderSide {
	if t.ruleStrategy.ShouldEnter(t.series.LastIndex(), t.record) {
		t.record.Operate(techan.Order{
			Side:          techan.BUY,
			Security:      "testid",
			Price:         t.series.LastCandle().ClosePrice,
			Amount:        big.NewDecimal(1),
			ExecutionTime: t.series.LastCandle().Period.End,
		})
		return BUY
	} else if t.ruleStrategy.ShouldExit(t.series.LastIndex(), t.record) {
		t.record.Operate(techan.Order{
			Side:          techan.SELL,
			Security:      "testid",
			Price:         t.series.LastCandle().ClosePrice,
			Amount:        big.NewDecimal(1),
			ExecutionTime: t.series.LastCandle().Period.End,
		})
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

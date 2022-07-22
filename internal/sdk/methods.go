package sdk

import (
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	pb "invest-bot/api/proto"
	"invest-bot/internal/config"
	"log"
	"os"
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

// PostSandboxOrder - выставление поручения в песочнице
func (s *SDK) PostSandboxOrder(figi string, quantity int64, direction pb.OrderDirection) (float64, float64, string, bool) {
	orderID := uuid.NewString()
	resp, err := s.Sandbox.PostSandboxOrder(s.Ctx, &pb.PostOrderRequest{
		Figi:      figi,
		Quantity:  quantity,
		Price:     nil,
		Direction: direction,
		AccountId: s.TradeConfig.AccountID,
		OrderType: pb.OrderType_ORDER_TYPE_MARKET,
		OrderId:   orderID,
	})
	if err != nil {
		log.Println("sandbox post order error:", err)
		return 0, 0, "", false
	}
	ok := resp.ExecutionReportStatus == pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL
	executedInstrumentPrice := s.MoneyValueToFloat(resp.ExecutedOrderPrice) // / float64(quantity)
	return executedInstrumentPrice, s.MoneyValueToFloat(resp.TotalOrderAmount), orderID, ok
}

// GetLotsByFigi - получение количества инструментов в 1 лоте
func (s *SDK) GetLotsByFigi(figi string) int64 {
	instrumentResp, err := s.Instruments.GetInstrumentBy(s.Ctx, &pb.InstrumentRequest{
		IdType:    pb.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI,
		ClassCode: "",
		Id:        figi,
	})
	if err != nil {
		log.Println("get lot by figi error", err)
	}
	return int64(instrumentResp.GetInstrument().Lot)
}

// CreateAccountIDs - если в конфиге нет id аккаунта песочницы/реального, то он откроется/запросится и запишется в конфиг
func (s *SDK) CreateAccountIDs(tc *config.TradeConfig) {
	if tc.AccountID == "" {
		fmt.Println("Sandbox account id field is empty in tradeconfig")
		openAccountReq, err := s.Sandbox.OpenSandboxAccount(s.Ctx, &pb.OpenSandboxAccountRequest{})
		if err != nil {
			log.Println("sandbox account opening error", err)
		}
		accountID := openAccountReq.AccountId
		tc.AccountID = accountID
		s.TradeConfig = tc
		cnf, err := yaml.Marshal(tc)
		if err != nil {
			log.Println(err)
		}
		err = os.WriteFile("config.yaml", cnf, 0666)
		if err != nil {
			log.Println(err)
		} else {
			fmt.Println("Account ID was successfully created")
		}
	}
	if tc.RealAccountID == "" {
		fmt.Println("Real account id field is empty in tradeconfig")
		getAccountsResp, err := s.Users.GetAccounts(s.Ctx, &pb.GetAccountsRequest{})
		if err != nil {
			log.Println("get user accounts error", err)
		}
		for _, acc := range getAccountsResp.GetAccounts() {
			if acc.AccessLevel == pb.AccessLevel_ACCOUNT_ACCESS_LEVEL_FULL_ACCESS {
				tc.RealAccountID = acc.GetId()
				break
			}
		}
		s.TradeConfig = tc
		cnf, err := yaml.Marshal(tc)
		if err != nil {
			log.Println(err)
		}
		err = os.WriteFile("config.yaml", cnf, 0666)
		if err != nil {
			log.Println(err)
		} else {
			fmt.Println("Real account ID was successfully saved")
		}
	}
}

// GetMoneyBalance - метод получения баланса счета реального/песочницы
func (s *SDK) GetMoneyBalance() int64 {
	var positionsResp *pb.PositionsResponse
	var err error
	if s.TradingMode == Sandbox {
		positionsResp, err = s.Sandbox.GetSandboxPositions(s.Ctx, &pb.PositionsRequest{AccountId: s.TradeConfig.AccountID})
		if err != nil {
			log.Println("get sandbox balance error", err)
		}
	} else if s.TradingMode == Real {
		positionsResp, err = s.Operations.GetPositions(s.Ctx, &pb.PositionsRequest{AccountId: s.TradeConfig.RealAccountID})
		if err != nil {
			log.Println("Getting real balance error", err)
		}
	}
	wasFound := false
	var balance int64 = 0
	for _, m := range positionsResp.GetMoney() {
		if m.Currency == "rub" {
			balance = m.GetUnits()
			wasFound = true
		}
	}
	if !wasFound {
		log.Println("currency not found in positions")
	}
	return balance
}

// PostMarketOrder - выставление заявки по рыночной цене на реальном счете
func (s *SDK) PostMarketOrder(figi string, quantity int64, direction pb.OrderDirection) (float64, float64, string, bool) {
	orderID := uuid.NewString()
	resp, err := s.Orders.PostOrder(s.Ctx, &pb.PostOrderRequest{
		Figi:      figi,
		Quantity:  quantity,
		Price:     nil,
		Direction: direction,
		AccountId: s.TradeConfig.RealAccountID,
		OrderType: pb.OrderType_ORDER_TYPE_MARKET,
		OrderId:   orderID,
	})
	if err != nil {
		log.Println("Market post order error:", err)
		return 0, 0, "", false
	}
	ok := resp.ExecutionReportStatus == pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL
	executedInstrumentPrice := s.MoneyValueToFloat(resp.ExecutedOrderPrice) // / float64(quantity)
	return executedInstrumentPrice, s.MoneyValueToFloat(resp.TotalOrderAmount), orderID, ok
}

// GetTickerByFigi - метод получения тикера по фиги инструмента
func (s *SDK) GetTickerByFigi(figi string) string {
	instrumentResp, err := s.Instruments.GetInstrumentBy(s.Ctx, &pb.InstrumentRequest{
		IdType:    pb.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI,
		ClassCode: "",
		Id:        figi,
	})
	if err != nil {
		log.Println("ticker getting error:", err)
	}
	return instrumentResp.GetInstrument().Ticker
}

// GetLastPrice - получение последней рыночной цены инструмента
func (s *SDK) GetLastPrice(figi string) float64 {
	priceResp, err := s.Marketdata.GetLastPrices(s.Ctx, &pb.GetLastPricesRequest{Figi: s.TradeConfig.TradeInstruments})
	if err != nil {
		log.Println("getting last price error", err)
	}
	lastPrices := make(map[string]float64, 0)
	for _, p := range priceResp.GetLastPrices() {
		lastPrices[p.Figi] = s.QuotationToFloat(p.Price)
	}
	if price, ok := lastPrices[figi]; ok {
		return price
	} else {
		log.Println("last price not found")
		return 0
	}
}

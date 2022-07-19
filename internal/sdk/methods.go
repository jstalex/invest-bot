package sdk

import (
	"fmt"
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
func (s *SDK) PostSandboxOrder(figi string, quantity int64, direction pb.OrderDirection) (float64, bool) {
	resp, err := s.Sandbox.PostSandboxOrder(s.Ctx, &pb.PostOrderRequest{
		Figi:      figi,
		Quantity:  quantity,
		Price:     nil,
		Direction: direction,
		AccountId: s.TradeConfig.AccountID,
		OrderType: pb.OrderType_ORDER_TYPE_MARKET,
		OrderId:   "testid",
	})
	if err != nil {
		log.Println("sandbox post order error:", err)
	}
	ok := resp.ExecutionReportStatus == pb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL
	return s.MoneyValueToFloat(resp.TotalOrderAmount), ok
}

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
func (s *SDK) GetSandboxMoneyBalance() int64 {
	positionsResp, err := s.Sandbox.GetSandboxPositions(s.Ctx, &pb.PositionsRequest{AccountId: s.TradeConfig.AccountID})
	if err != nil {
		log.Println("get sandbox balance error", err)
	}
	wasFound := false
	var balance int64 = 0
	for _, m := range positionsResp.GetMoney() {
		if m.Currency == "rub" {
			balance = m.Units
			wasFound = true
		}
	}
	if !wasFound {
		log.Println("currency not found in positions")
	}
	return balance
}
func (s *SDK) CreateSandboxAccount(tc *config.TradeConfig) {
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
		}
		fmt.Println("Account ID was successfully created")
	}
}

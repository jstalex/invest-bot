package sdk

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	pb "invest-bot/api/proto"
	"log"
	"math"
	"time"
)

const shortForm = "2006-04-02"

func (s *SDK) DateToTimestamp(in string) *timestamp.Timestamp {
	t, err := time.Parse(shortForm, in)
	if err != nil {
		log.Println(err)
	}
	return &timestamp.Timestamp{Seconds: t.Unix()}
}

func (s *SDK) HistoricalCandleToTechan(hc *pb.HistoricCandle) *techan.Candle {
	minutes := 1
	switch s.TradeConfig.Period {
	case 1:
		minutes = 1
	case 2:
		minutes = 5
	case 3:
		minutes = 15
	case 4:
		minutes = 60
	}
	period := techan.NewTimePeriod(hc.Time.AsTime(), time.Duration(minutes)*time.Minute)
	candle := techan.NewCandle(period)
	candle.OpenPrice = big.NewDecimal(s.QuotationToFloat(hc.Open))
	candle.ClosePrice = big.NewDecimal(s.QuotationToFloat(hc.Close))
	candle.MaxPrice = big.NewDecimal(s.QuotationToFloat(hc.High))
	candle.MinPrice = big.NewDecimal(s.QuotationToFloat(hc.Low))
	return candle
}

func (s *SDK) QuotationToFloat(src *pb.Quotation) float64 {
	return float64(src.Units) + float64(src.Nano)*math.Pow10(-9)
}

func (s *SDK) MoneyValueToFloat(src *pb.MoneyValue) float64 {
	return float64(src.Units) + float64(src.Nano)*math.Pow10(-9)
}

func (s *SDK) PBCandleToTechan(pbc *pb.Candle) *techan.Candle {
	minutes := 1
	switch s.TradeConfig.Period {
	case 1:
		minutes = 1
	case 2:
		minutes = 5
	case 3:
		minutes = 15
	case 4:
		minutes = 60
	}
	period := techan.NewTimePeriod(pbc.Time.AsTime(), time.Duration(minutes)*time.Minute)
	candle := techan.NewCandle(period)
	candle.OpenPrice = big.NewDecimal(s.QuotationToFloat(pbc.Open))
	candle.ClosePrice = big.NewDecimal(s.QuotationToFloat(pbc.Close))
	candle.MaxPrice = big.NewDecimal(s.QuotationToFloat(pbc.High))
	candle.MinPrice = big.NewDecimal(s.QuotationToFloat(pbc.Low))
	return candle
}

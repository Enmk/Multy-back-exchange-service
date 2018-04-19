package core

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Appscrunch/Multy-back-exchange-service/api"
	"github.com/Appscrunch/Multy-back-exchange-service/currencies"
)

type BitfinexManager struct {
	bitfinexTickers map[int]BitfinexTicker
	api             *api.BitfinexApi
}

type BitfinexTicker struct {
	ChanID     int    `json:"chanId"`
	Channel    string `json:"channel"`
	Event      string `json:"event"`
	Pair       string `json:"pair"`
	Symbol     string `json:"symbol"`
	Rate       string
	TimpeStamp time.Time
}

func (bitfinexTicker BitfinexTicker) IsFilled() bool {
	return (len(bitfinexTicker.Symbol) > 0 && len(bitfinexTicker.Rate) > 0)
}

func (b *BitfinexTicker) getCurriences() (currencies.Currency, currencies.Currency) {
	//fmt.Println(b.Symbol)
	if len(b.Symbol) > 0 {
		var symbol = b.Symbol
		var damagedSymbol = TrimLeftChars(symbol, 2)
		for _, referenceCurrency := range currencies.DefaultReferenceCurrencies {
			//fmt.Println(damagedSymbol, referenceCurrency.CurrencyCode())

			referenceCurrencyCode := referenceCurrency.CurrencyCode()

			if referenceCurrencyCode == "USDT" {
				referenceCurrencyCode = "USD"
			}

			//fmt.Println(damagedSymbol)
			//fmt.Println(referenceCurrencyCode)
			//fmt.Println(strings.Contains(damagedSymbol, referenceCurrencyCode))

			if strings.Contains(damagedSymbol, referenceCurrencyCode) {
				//fmt.Println(damagedSymbol)

				//fmt.Println("2",symbol, referenceCurrency.CurrencyCode())
				targetCurrencyStringWithT := strings.TrimSuffix(symbol, referenceCurrencyCode)
				targetCurrencyString := TrimLeftChars(targetCurrencyStringWithT, 1)
				//fmt.Println("targetCurrencyString", targetCurrencyString)
				var targetCurrency = currencies.NewCurrencyWithCode(targetCurrencyString)
				return targetCurrency, referenceCurrency
			}
		}

	}
	return currencies.NotAplicable, currencies.NotAplicable
}

func (b *BitfinexManager) StartListen(exchangeConfiguration ExchangeConfiguration, callback func(tickerCollection TickerCollection, err error)) {
	b.bitfinexTickers = make(map[int]BitfinexTicker)
	b.api = api.NewBitfinexApi()

	var apiCurrenciesConfiguration = api.ApiCurrenciesConfiguration{}
	apiCurrenciesConfiguration.TargetCurrencies = exchangeConfiguration.TargetCurrencies
	apiCurrenciesConfiguration.ReferenceCurrencies = exchangeConfiguration.ReferenceCurrencies

	go b.api.StartListen(apiCurrenciesConfiguration, func(message []byte, err error) {
		//fmt.Println(0)
		if err != nil {
			log.Println("error:", err)
			callback(TickerCollection{}, err)
		} else if message != nil {
			//fmt.Printf("%s \n", message)
			//fmt.Println(1)
			b.addMessage(message)
			//fmt.
		} else {
			fmt.Println("error parsing Bitfinex ticker:", err)
		}
	})

	for range time.Tick(1 * time.Second) {
		//TODO: add check if data is old and don't sent it to callback
		func() {
			tickers := []Ticker{}
			for _, value := range b.bitfinexTickers {
				if value.TimpeStamp.After(time.Now().Add(-maxTickerAge * time.Second)) {
					var ticker = Ticker{}
					ticker.Rate = value.Rate
					ticker.Symbol = value.Symbol

					targetCurrency, referenceCurrency := value.getCurriences()
					ticker.TargetCurrency = targetCurrency
					ticker.ReferenceCurrency = referenceCurrency
					tickers = append(tickers, ticker)
				}
			}

			var tickerCollection = TickerCollection{}
			tickerCollection.TimpeStamp = time.Now()
			tickerCollection.Tickers = tickers
			//fmt.Println(tickerCollection)
			if len(tickerCollection.Tickers) > 0 {
				callback(tickerCollection, nil)
			}
		}()
	}
}

func (b *BitfinexManager) addMessage(message []byte) {

	var bitfinexTicker BitfinexTicker
	json.Unmarshal(message, &bitfinexTicker)

	if bitfinexTicker.ChanID > 0 {
		//fmt.Println(bitfinexTicker)
		b.bitfinexTickers[bitfinexTicker.ChanID] = bitfinexTicker
	} else {

		var unmarshaledTickerMessage []interface{}
		json.Unmarshal(message, &unmarshaledTickerMessage)

		if len(unmarshaledTickerMessage) > 1 {
			var chanId = int(unmarshaledTickerMessage[0].(float64))
			//var unmarshaledTicker []interface{}
			if v, ok := unmarshaledTickerMessage[1].([]interface{}); ok {
				var sub = b.bitfinexTickers[chanId]
				sub.Rate = strconv.FormatFloat(v[0].(float64), 'f', 8, 64)
				sub.TimpeStamp = time.Now()
				b.bitfinexTickers[chanId] = sub
			}
		}
	}
}

//func (b PoloniexManager) convertArgsToTicker(args []interface{}) (wsticker PoloniexTicker, err error) {
//	wsticker.CurrencyPair = b.channelsByID[strconv.FormatFloat(args[0].(float64), 'f', 0, 64)]
//	wsticker.Last = args[1].(string)
//	return
//}
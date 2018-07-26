/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/

package server

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Appscrunch/Multy-back-exchange-service/currencies"
	"github.com/Appscrunch/Multy-back-exchange-service/exchange-rates"

	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"

	_ "github.com/KristinaEtc/slflog"
)

type MockDataProvider struct {
	Tickers []exchangeRates.Ticker
}

func NewMockDataProvider() *MockDataProvider {
	result := &MockDataProvider{[]exchangeRates.Ticker{
		exchangeRates.Ticker{
			Pair: currencies.CurrencyPair{
				TargetCurrency:    0,
				ReferenceCurrency: 5,
			},
			Rate:         1.0,
			TimpeStamp:   time.Now(),
			IsCalculated: false,
		},
		exchangeRates.Ticker{
			Pair: currencies.CurrencyPair{
				TargetCurrency:    6,
				ReferenceCurrency: 7,
			},
			Rate:         1.0,
			TimpeStamp:   time.Now(),
			IsCalculated: false,
		},
		exchangeRates.Ticker{
			Pair: currencies.CurrencyPair{
				TargetCurrency:    3,
				ReferenceCurrency: -3,
			},
			Rate:         99.3,
			TimpeStamp:   time.Now(),
			IsCalculated: false,
		},
	},
	}

	return result
}

func (self *MockDataProvider) GetRates(timeStamp time.Time, exchangeName string, targetCode string, referecies []string) ([]exchangeRates.Ticker, error) {
	return append(make([]exchangeRates.Ticker, 0, len(self.Tickers)), self.Tickers...), nil
}

func (self *MockDataProvider) GetHistoryRates(from time.Time, to time.Time, exchangeName string, targetCode string, referecies []string) ([]exchangeRates.Ticker, error) {
	return append(make([]exchangeRates.Ticker, 0, len(self.Tickers)), self.Tickers...), nil
}

func TestSocketIo(t *testing.T) {
	mockDataProvider := NewMockDataProvider()
	config := MarketSocketIoServerConfig{
		"localhost",
		8088,
	}
	marketServer := NewMarketSocketIoServer(config)

	go marketServer.Start(mockDataProvider)
	defer marketServer.Stop()

	serverURL := gosocketio.GetUrl(config.Host, config.Port, false)
	t.Logf("Dialing server on %s ...", serverURL)

	marketClient, err := gosocketio.Dial(serverURL, transport.GetDefaultWebsocketTransport())
	if err != nil {
		t.Errorf("Can't dial to server: %v", err)
		return
	}

	strResponse, err := marketClient.Ack("/get_rates", marketRateRequest{
		"exchange",
		time.Now(),
		"FOO",
		[]string{"abc", "def"},
	},
		1*time.Second)

	if err != nil {
		t.Errorf("Failed to receive a response on 'get_rates' request: %v", err)
		return
	}

	var response marketRateResponse
	err = json.Unmarshal([]byte(strResponse), &response)

	if err != nil {
		t.Errorf("Failed to unmarshal response as MarketRateResponse: %v\n%s", err, strResponse)
		return
	}

	if len(response.Rates) != len(mockDataProvider.Tickers) {
		t.Errorf("Invalid response size: expected %v got %v", len(response.Rates), len(mockDataProvider.Tickers))
		return
	}

	//
	strResponse, err = marketClient.Ack("/get_rates_history", marketRateHistoryRequest{
		"exchange",
		time.Now().Add(time.Duration(-11000) * time.Second),
		time.Now(),
		"FOO",
		[]string{"abc", "def"},
	},
		1*time.Second)

	if err != nil {
		t.Errorf("Failed to receive a response on 'get_rates_history' request: %v", err)
		return
	}

	var history marketRateHistoryResponse
	err = json.Unmarshal([]byte(strResponse), &history)
	t.Logf("Got response: %s, %v", strResponse, history)

	if err != nil {
		t.Errorf("Failed to unmarshal response as MarketRateResponse: %v\n%s", err, strResponse)
		return
	}

	// if len(response.Rates) != len(mockDataProvider.Tickers) {
	// 	t.Errorf("Invalid response size: expected %v got %v", len(history.history), len(mockDataProvider.Tickers))
	// 	return
	// }
}

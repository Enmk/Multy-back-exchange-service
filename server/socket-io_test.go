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

	"github.com/Appscrunch/Multy-back-exchange-service/exchange-rates"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

type MockDataProvider struct {
	Tickers []exchangeRates.Ticker
}

func NewMockDataProvider() *MockDataProvider {
	return &MockDataProvider{make([]exchangeRates.Ticker, 0, 10)}
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

	serverURL := gosocketio.GetUrl(config.Host, config.Port, false)
	t.Logf("Dialing server on %s ...", serverURL)

	marketClient, err := gosocketio.Dial(serverURL, transport.GetDefaultWebsocketTransport())
	if err != nil {
		t.Errorf("Can't dial to server: %v", err)
		return
	}

	strResponse, err := marketClient.Ack("/get_rates", MarketRateRequest{
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

	var response MarketRateResponse
	err = json.Unmarshal([]byte(strResponse), response)

	if err != nil {
		t.Errorf("Failed to unmarshal response as MarketRateResponse: %v\n%s", err, strResponse)
		return
	}
}

/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"

	"github.com/KristinaEtc/slf"

	"github.com/Appscrunch/Multy-back-exchange-service/exchange-rates"
)

var log = slf.WithContext("Market-SocketIoServer")

type SocketIoConfig struct {
	address string
}

type rateData map[string]float64
type rateHistory map[time.Time]rateData

type marketRateRequest struct {
	Exchange   string    `json:"exchange" binding:"required"`
	Date       time.Time `json:"from" time_format:"RFC3339" binding:"required"`
	Reference  string    `json:"reference_currency_id" binding:"required"`
	Currencies []string  `json:"currency_ids" binding:"required"`
}

type marketRateResponse struct {
	Error error `json:"error"`
	// exchange  string    `json:"exchange" binding:"required"`
	// date      time.Time `json:"date" time_format:"RFC3339" binding:"required"`
	// reference string    `json:"reference_currency_id" binding:"required"`
	Rates rateData `json:"rates" binding:"required"`
}

type marketRateHistoryRequest struct {
	Exchange   string    `json:"exchange" binding:"required"`
	From       time.Time `json:"from" time_format:"RFC3339" binding:"required"`
	To         time.Time `json:"to" time_format:"RFC3339" binding:"required"`
	Reference  string    `json:"reference_currency_id" binding:"required"`
	Currencies []string  `json:"currency_ids" binding:"required"`
}

type marketRateHistoryResponse struct {
	Error error `json:"error"`
	// exchange  string      `json:"exchange" binding:"required"`
	// from      time.Time   `json:"date" time_format:"RFC3339" binding:"required"`
	// to        time.Time   `json:"to" time_format:"RFC3339" binding:"required"`
	// reference string      `json:"reference_currency_id" binding:"required"`
	History rateHistory `json:"history" binding:"required"`
}

type MarketSocketIoServerConfig struct {
	Host string
	Port int
}
type MarketSocketIoServer struct {
	// TODO: add logging
	serveMux *http.ServeMux
	server   *http.Server
	config   MarketSocketIoServerConfig
}

func NewMarketSocketIoServer(config MarketSocketIoServerConfig) *MarketSocketIoServer {
	return &MarketSocketIoServer{nil, nil, config}
}

func (self *MarketSocketIoServer) Start(dataProvider exchangeRates.MarketDataProvider) error {
	handler := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())

	handler.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		log.Debugf("Connected: %v from %v", c.Id(), c.Ip())
	})

	handler.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		log.Debugf("Disconnected: %v", c.Id())
	})

	handler.On("/get_rates", func(c *gosocketio.Channel, request marketRateRequest) marketRateResponse {
		log = log.WithField("request:", "get_rates")

		rawRates, err := dataProvider.GetRates(request.Date, request.Exchange, request.Reference, request.Currencies)
		if err != nil {
			log.WithError(err).Errorf("Failed to get rates with %v", request)
			return marketRateResponse{
				Error: err,
			}
		}

		rates := make(rateData)
		for _, ticker := range rawRates {
			rates[strconv.Itoa(int(ticker.Pair.TargetCurrency))] = ticker.Rate
		}

		log.Debugf("Resulting rates count: %v", len(rates))
		return marketRateResponse{
			Rates: rates,
		}
	})

	handler.On("/get_rates_history", func(c *gosocketio.Channel, request marketRateHistoryRequest) marketRateHistoryResponse {
		log = log.WithField("request:", "get_rates_history")

		// TODO: choose granularity based on From - To duration
		rawHistory, err := dataProvider.GetHistoryRates(request.From, request.To, request.Exchange, request.Reference, request.Currencies)
		if err != nil {
			log.WithError(err).Errorf("Failed to get rates history with %v", request)
			return marketRateHistoryResponse{
				Error: err,
			}
		}

		history := make(rateHistory)
		for _, ticker := range rawHistory {
			// TODO: trim timestamp according to granularity to force more values in same bucket.
			rates := history[ticker.TimpeStamp]
			if rates == nil {
				history[ticker.TimpeStamp] = make(rateData)
				rates = history[ticker.TimpeStamp]
			}
			rates[strconv.Itoa(int(ticker.Pair.TargetCurrency))] = ticker.Rate
		}

		log.Debugf("Resulting history size: %v", len(history))
		return marketRateHistoryResponse{
			History: history,
		}
	})

	self.serveMux = http.NewServeMux()
	self.serveMux.Handle("/socket.io", handler)

	addr := self.config.Host + ":" + strconv.Itoa(self.config.Port)
	self.server = &http.Server{Addr: addr, Handler: handler}

	return self.server.ListenAndServe()
}

// Stops the server
func (self *MarketSocketIoServer) Stop() {
	self.server.Close()
}

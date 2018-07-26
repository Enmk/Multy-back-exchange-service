/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package server

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"

	"github.com/Appscrunch/Multy-back-exchange-service/exchange-rates"
)

type SocketIoConfig {
	address string
}

type rateData map[string]float64
type rateHistory map[time.Time]rateData

type marketResponse struct {
	Error *error `json:"error"`
}

type marketRateRequest struct {
	Exchange   string    `json:"exchange" binding:"required"`
	Date       time.Time `json:"from" time_format:"RFC3339"`
	Reference  string    `json:"reference_currency_id" binding:"required"`
	Currencies []string  `json:"currency_ids" binding:"required"`
}

type marketRateResponse struct {
	marketResponse
	// exchange  string    `json:"exchange" binding:"required"`
	// date      time.Time `json:"date" time_format:"RFC3339" binding:"required"`
	// reference string    `json:"reference_currency_id" binding:"required"`
	Rates rateData `json:"rates" binding:"required"`
}

type marketRateHistoryRequest struct {
	Exchange   string    `json:"exchange" binding:"required"`
	From       time.Time `json:"from" time_format:"RFC3339"`
	To         time.Time `json:"to" time_format:"RFC3339"`
	Reference  string    `json:"reference_currency_id" binding:"required"`
	Currencies []string  `json:"currency_ids" binding:"required"`
}

type marketRateHistoryResponse struct {
	marketResponse
	// exchange  string      `json:"exchange" binding:"required"`
	// from      time.Time   `json:"date" time_format:"RFC3339" binding:"required"`
	// to        time.Time   `json:"to" time_format:"RFC3339" binding:"required"`
	// reference string      `json:"reference_currency_id" binding:"required"`
	History rateHistory `json:"history" binding:"required"`
}

func ServeSocketIo(exchangeManager *exchangeRates.ExchangeManager) {
	server := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())

	server.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		log.Printf("Connected: %v from ", c.Id(), c.Ip())

		// c.Emit("/message", Message{10, "main", "using emit"})

		// c.Join("test")
		// c.BroadcastTo("test", "/message", Message{10, "main", "using broadcast"})
	})

	server.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		log.Printf("Disconnected: %v", c.Id())
	})

	server.On("/get_market_rate", func(c *gosocketio.Channel, request marketRateRequest) marketRateResponse {
		rawRates, err := exchangeManager.GetRates(date, exchange, reference, currencyIds)
		if err != nil {
			return exchangeRateResponse {
				Error: err,
			}
		}

		var rates rateData
		for _, ticker := range rawRates {
			rates[strconv.Itoa(int(ticker.Pair.TargetCurrency))] = ticker.Rate
		}

		return exchangeRateResponse {
			// Exchange:  exchange,
			// Date:      date,
			// Reference: reference,
			Rates: rates,
		}
	})

	server.On("/get_market_rate", func(c *gosocketio.Channel, request marketRateHistoryRequest) marketRateResponse {
		rawHistory, err := self.exchangeManager.GetHistoryRates(request.from, request.to, request.exchange, request.reference, request.currencies)
		if err != nil {
			return marketRateResponse {
				Error: err
			}
		}

		var history rateHistory
		for _, ticker := range rawHistory {
			rates := history[ticker.TimpeStamp]
			if rates == nil {
				history[ticker.TimpeStamp] = rateData{}
				rates = history[ticker.TimpeStamp]
			}
			rates[strconv.Itoa(int(ticker.Pair.TargetCurrency))] = ticker.Rate
		}

		return marketRateHistoryResponse {
			// Exchange:  exchange,
			// Date:      date,
			// Reference: reference,
			History: history
		}
	})

	serveMux := http.NewServeMux()
	serveMux.Handle("/socket.io/v1", server)

	log.Panic(http.ListenAndServe(":3811", serveMux))
}

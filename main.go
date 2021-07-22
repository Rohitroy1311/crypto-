package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"

	"net/http"

	"github.com/gorilla/websocket"
)

type currencyDetails struct {
	Id          string `json:"id"`
	FullName    string `json:"fullName"`
	Ask         string `json:"ask"`
	Bid         string `json:"bid"`
	Last        string `json:"last"`
	Open        string `json:"open"`
	Low         string `json:"low"`
	High        string `json:"high"`
	FeeCurrency string `json:"feeCurrency"`
}

type symbolTickerParams struct {
	Ask         string `json:"ask"`
	Bid         string `json:"bid"`
	Last        string `json:"last"`
	Open        string `json:"open"`
	Low         string `json:"low"`
	High        string `json:"high"`
	Volume      string `json:"volume"`
	VolumeQuote string `json:"volumeQuote"`
	Timestamp   string `json:"timestamp"`
	Symbol      string `json:"symbol"`
}

type symbolTickerResponse struct {
	Jsonrpc string             `json:"jsonrpc"`
	Method  string             `json:"method"`
	Params  symbolTickerParams `json:"params"`
}

type currencyArrayDetails struct {
	Currencies [2]currencyDetails `json:"currencies"`
}

var addr = flag.String("addr", "api.hitbtc.com", "http service address")

var currencyBTCUSD, currencyETHBCD currencyDetails
var currencyAll currencyArrayDetails

func endpointBTCUSD(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request URL: %s", r.URL)
	response, err := json.Marshal(currencyBTCUSD)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Printf("Response: %s", string(response))
	w.Write(response)
}

func endpointETHBCD(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request URL: %s", r.URL)
	response, err := json.Marshal(currencyETHBCD)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Printf("Response: %s", string(response))
	w.Write(response)
}

func endpointAll(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request URL: %s", r.URL)
	response, err := json.Marshal(currencyAll)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Printf("Response: %s", string(response))
	w.Write(response)
}

func handleRequests() {
	http.HandleFunc("/currency/BTCUSD", endpointBTCUSD)
	http.HandleFunc("/currency/ETHBCD", endpointETHBCD)
	http.HandleFunc("/currency/all", endpointAll)
	log.Fatal(http.ListenAndServe(":12121", nil))
}

func syncTickerData() {
	flag.Parse()
	log.SetFlags(0)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "api/2/ws"}
	// log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	subscribeBTCUSDReq := "{\"method\": \"subscribeTicker\",\"params\": { \"symbol\": \"BTCUSD\"},\"id\": 3333}"
	err = c.WriteMessage(websocket.TextMessage, []byte(subscribeBTCUSDReq))
	if err != nil {
		log.Println("write:", err)
		return
	}

	subscribeETHBTCReq := "{\"method\": \"subscribeTicker\",\"params\": { \"symbol\": \"ETHBTC\"},\"id\": 4444}"
	err = c.WriteMessage(websocket.TextMessage, []byte(subscribeETHBTCReq))
	if err != nil {
		log.Println("write:", err)
		return
	}

	currencyBTCUSD.Id = "BTC"
	currencyBTCUSD.FullName = "Bitcoin"
	currencyBTCUSD.FeeCurrency = "USD"

	currencyETHBCD.Id = "ETH"
	currencyETHBCD.FullName = "Ethereum"
	currencyETHBCD.FeeCurrency = "BCD"

	done := make(chan struct{})

	go func() {
		var stTickerResponse symbolTickerResponse
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				continue
			}

			// log.Printf("recv: %s", message)

			err = json.Unmarshal(message, &stTickerResponse)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if stTickerResponse.Params.Symbol == "BTCUSD" {
				currencyBTCUSD.Ask = stTickerResponse.Params.Ask
				currencyBTCUSD.Bid = stTickerResponse.Params.Bid
				currencyBTCUSD.Last = stTickerResponse.Params.Last
				currencyBTCUSD.Open = stTickerResponse.Params.Open
				currencyBTCUSD.Low = stTickerResponse.Params.Low
				currencyBTCUSD.High = stTickerResponse.Params.High
				currencyAll.Currencies[0] = currencyBTCUSD
			} else if stTickerResponse.Params.Symbol == "ETHBTC" {
				currencyETHBCD.Ask = stTickerResponse.Params.Ask
				currencyETHBCD.Bid = stTickerResponse.Params.Bid
				currencyETHBCD.Last = stTickerResponse.Params.Last
				currencyETHBCD.Open = stTickerResponse.Params.Open
				currencyETHBCD.Low = stTickerResponse.Params.Low
				currencyETHBCD.High = stTickerResponse.Params.High
				currencyAll.Currencies[1] = currencyETHBCD
			}
		}
	}()

	// ticker := time.NewTicker(3 * time.Second)
	// defer ticker.Stop()

	for {
		select {
		case <-done:
			return
			// case t := <-ticker.C:
			// case <-ticker.C:

		}
	}
}
func main() {
	go handleRequests()
	syncTickerData()
}

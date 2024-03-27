package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TON struct {
	Status struct {
		Timestamp    time.Time `json:"timestamp"`
		ErrorCode    int       `json:"error_code"`
		ErrorMessage string    `json:"error_message"`
		Elapsed      int       `json:"elapsed"`
		CreditCount  int       `json:"credit_count"`
		Notice       string    `json:"notice"`
	} `json:"status"`
	Data struct {
		TON struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
			Quote  struct {
				USD struct {
					Price       float64   `json:"price"`
					LastUpdated time.Time `json:"last_updated"`
				} `json:"USD"`
			} `json:"quote"`
		} `json:"TON"`
	} `json:"data"`
}

func main() {
	bot, err := tgbotapi.NewBotAPI("tgBotToken")
	if err != nil {
		log.Panic(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=TON&convert=USD", nil)
	if err != nil {
		log.Print(err)
	}

	req.Header.Add("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", "yourApiToken")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	var lastCommandTimestamp = make(map[int]int64)
	go priceChangeRoutine(bot, client, req)

	for update := range updates {
		resp, err := client.Do(req)
		if err != nil {
			log.Print(err)
		}
		var response TON
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			log.Print(err)
		}
		price := response.Data.TON.Quote.USD.Price
		if update.Message.Command() == "price" {
			userID := update.Message.From.ID
			lastTimestamp, exists := lastCommandTimestamp[userID]
			if exists && time.Now().Unix()-lastTimestamp < 10 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please wait at least 10 seconds before using the /price command again.")
				_, err = bot.Send(msg)
				if err != nil {
					log.Print(err)
				}
			} else {
				lastCommandTimestamp[userID] = time.Now().Unix()
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Current TON price: $%.5f", price))
				_, err = bot.Send(msg)
				if err != nil {
					log.Print(err)
				}
			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command.")
			_, err = bot.Send(msg)
			if err != nil {
				log.Print(err)
			}
			continue
		}
	}
}
func priceChangeRoutine(bot *tgbotapi.BotAPI, client *http.Client, req *http.Request) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	var previousPrice float64

	for range ticker.C {
		resp, err := client.Do(req)
		if err != nil {
			log.Print(err)
			continue
		}

		var response TON
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			log.Print(err)
			continue
		}

		newPrice := response.Data.TON.Quote.USD.Price

		if previousPrice != 0 && math.Abs(newPrice-previousPrice) >= 0.01 {
			msg := tgbotapi.NewMessage(1234, fmt.Sprintf("Current TON price: $%.5f", newPrice)) //change 1234 to your ChatID
			_, err := bot.Send(msg)
			if err != nil {
				log.Print(err)
				continue
			}
		}
		previousPrice = newPrice
	}
}

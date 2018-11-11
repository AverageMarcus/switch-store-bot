package main

import (
	"encoding/json"

	"gopkg.in/resty.v1"
)

type Game struct {
	Title                   string  `json:"title"`
	PriceDiscountPercentage float32 `json:"price_discount_percentage_f"`
	Price                   float32 `json:"price_lowest_f"`
	ID                      string  `json:"fs_id"`
	ChangeDate              string  `json:"change_date"`
	URL                     string  `json:"url"`
}

func FetchGames() []Game {
	resp, err := resty.R().
		SetQueryParams(map[string]string{
			"fq":    "type:GAME%20AND system_type:nintendoswitch*%20AND product_code_txt:* AND price_has_discount_b:true",
			"q":     "*",
			"rows":  "9999",
			"sort":  "sorting_title asc",
			"start": "0",
			"wt":    "json",
		}).Get("http://search.nintendo-europe.com/en/select")

	if err != nil {
		panic("Failed to get games")
	}

	var objmap map[string]*json.RawMessage
	err = json.Unmarshal(resp.Body(), &objmap)
	var response map[string]*json.RawMessage
	err = json.Unmarshal(*objmap["response"], &response)

	var games []Game
	if err := json.Unmarshal(*response["docs"], &games); err != nil {
		panic(err)
	}

	return games
}

func FindGame(query string) []Game {
	resp, err := resty.R().
		SetQueryParams(map[string]string{
			"fq":    "type:GAME%20AND system_type:nintendoswitch*%20AND product_code_txt:* AND title:*" + query + "*",
			"q":     "*",
			"rows":  "9999",
			"sort":  "sorting_title asc",
			"start": "0",
			"wt":    "json",
		}).Get("http://search.nintendo-europe.com/en/select")

	if err != nil {
		return nil
	}

	var objmap map[string]*json.RawMessage
	err = json.Unmarshal(resp.Body(), &objmap)
	var response map[string]*json.RawMessage
	err = json.Unmarshal(*objmap["response"], &response)

	var games []Game
	if err := json.Unmarshal(*response["docs"], &games); err != nil {
		panic(err)
	}

	return games
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"slices"
	"time"

	"github.com/valyala/fasthttp"
)

type Config struct {
	Hakusana string `json:"hakusana"`
	Webhook  string `json:"webhook"`
}

type Product struct {
	Heading      string   `json:"heading"`
	Location     string   `json:"location"`
	ImageURLs    []string `json:"image_urls"`
	Timestamp    int64    `json:"timestamp"`
	CanonicalURL string   `json:"canonical_url"`
	Price        struct {
		Amount       int    `json:"amount"`
		CurrencyCode string `json:"currency_code"`
		PriceUnit    string `json:"price_unit"`
	} `json:"price"`
}

type Response struct {
	Docs []Product `json:"docs"`
}

var foundTotal int

func main() {
	content, _ := ioutil.ReadFile("config.json")
	var config Config
	_ = json.Unmarshal(content, &config)
	hakusana := config.Hakusana
	webhook := config.Webhook

	seen := []string{}
	firstRun := true

	client := &fasthttp.Client{}

	for {
		foundTotal = 0
		req := fasthttp.AcquireRequest()
		req.SetRequestURI("https://beta.tori.fi/recommerce-search-page/api/search/SEARCH_ID_BAP_COMMON?q=" + hakusana + "&sort=PUBLISHED_DESC")
		req.Header.SetMethod("GET")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
		req.Header.Set("Accept", "application/json")

		resp := fasthttp.AcquireResponse()
		err := client.Do(req, resp)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			continue
		}

		var response Response
		body := resp.Body()
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			continue
		}

		for i, product := range response.Docs {
			if i >= 25 {
				break
			}
			if !slices.Contains(seen, product.CanonicalURL) {
				if !firstRun {
					foundTotal++
					embeds := "\"embeds\": ["
					for i := 0; i < len(product.ImageURLs) && i < 3; i++ {
						embeds += fmt.Sprintf(`{"title": "%s","url": "%s","description":"Hinta: %d %s","color": 2895667,"image": {"url": "%s"}}`,
							product.Heading, product.CanonicalURL, product.Price.Amount, product.Price.PriceUnit, product.ImageURLs[i])
						if i < len(product.ImageURLs)-1 && i < 2 {
							embeds += ","
						}
					}
					embeds += "]"
					payload := fmt.Sprintf(`{"content": null,%s,"attachments": []}`, embeds)
					sendWebhook(webhook, payload)
				}
				seen = append(seen, product.CanonicalURL)
			}
		}

		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)

		if foundTotal > 0 {
			fmt.Println("Uusia ilmoituksia löydetty: ", foundTotal)
		}

		firstRun = false
		time.Sleep(10 * time.Second)
	}
}

func sendWebhook(webhook, payload string) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(webhook)
	req.Header.SetMethod("POST")
	req.Header.Set("Content-Type", "application/json")
	req.SetBody([]byte(payload))

	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	err := client.Do(req, resp)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(resp)
}

package main

import (
	"bytes"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

func main() {
	hakusana := "" // HAKUSANA TÄHÄN
	webhook := ""  // WEBHOOK TÄHÄN

	seen := []string{}
	firstRun := true

	c := colly.NewCollector(colly.AllowURLRevisit())

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	})

	c.OnHTML("a.item_row_flex", func(h *colly.HTMLElement) {
		found := h.Request.AbsoluteURL(h.Attr("href"))
		imgUrl := ""
		h.ForEach("img[src]", func(_ int, img *colly.HTMLElement) {
			imgUrl = img.Attr("src")
			ruleIndex := strings.Index(imgUrl, "?rule=")
			if ruleIndex != -1 {
				imgUrl = imgUrl[:ruleIndex+6] + "medium_660"
			}
		})
		if !slices.Contains(seen, found) {
			if !firstRun {
				fmt.Println(found)
				sendWebhook(webhook, `{"content": null,"embeds": [{"title": "Uusi tuote löydetty 🔍","url": "`+found+`","description":"Hakusana: `+hakusana+`","color": 2895667,"image": {"url": "`+imgUrl+`"}}],"attachments": []}`)
			}
			seen = append(seen, found)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	})

	for {
		c.Visit("https://www.tori.fi/koko_suomi?q=" + hakusana + "&cg=0&w=3&st=s&ca=18&l=0&md=th")
		time.Sleep(10 * time.Second)
		firstRun = false
	}
}

func sendWebhook(webhook, payload string) {
	req, _ := http.NewRequest("POST", webhook, bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	client.Do(req)
}

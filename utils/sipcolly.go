package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gocolly/colly"
	// "github.com/gocolly/colly/proxy"
)



func sips() {
	fName := "bk_boxscores.csv"
	file, err := os.Create(fName)

	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{"URL"})

	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("basketball-reference.com", "www.basketball-reference.com", "www.bovada.lv"),
		// colly.MaxDepth(2),
		// colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	// detailCollector := c.Clone()

	// c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})

	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
		// fmt.Println(r.Headers)
		// fmt.Println(r.ID)
		// fmt.Println(r.Marshal)
		// fmt.Println(r.Ctx)
	})
	// fmt.Println("this is C: ", c.String())
	// c.OnHTML("body", func(e *colly.HTMLElement) {
	// 	fmt.Println("e.Text: ", e.Text)
	// })
	// // On every a HTML element which has name attribute call callback
	// c.OnHTML(`a[id]`, func(e *colly.HTMLElement) {
	// 	// Activate detailCollector if the link contains "coursera.org/learn"
	// 	courseURL := e.Request.AbsoluteURL(e.Attr("href"))
	// 	writer.Write([]string{
	// 		e.Request.AbsoluteURL(e.ChildAttr("a", "href")),
	// 	})
	// 	if strings.Index(courseURL, "basketball-reference.com/boxscores/") != -1 {
	// 		detailCollector.Visit(courseURL)
	// 	}
	// })

	// c.OnHTML("p[class=links]", func(e *colly.HTMLElement) {
	// 	fmt.Println(e.Request.AbsoluteURL(e.ChildAttr("a", "href")))
	// 	writer.Write([]string{
	// 		e.Request.AbsoluteURL(e.ChildAttr("a", "href")),
	// 	})
	// 	fmt.Println(e.Request.AbsoluteURL(e.ChildAttr("a", "href")))
	// })

	c.OnHTML("body", func(e *colly.HTMLElement) {
		fmt.Println("in pre")
		fmt.Println("dom children: ", e.DOM.Children().Text())
		jsonData := e.Text
		fmt.Println("jsonData: ", err)
		parsed := &ResultStruct{}
		err := json.Unmarshal([]byte(jsonData), parsed)
		fmt.Println("parsed: ", err)

		if err != nil {
			log.Fatal(err)
		}
	})

	d := c.Clone()
	d.OnResponse(func(r *colly.Response) {
		idStart := bytes.Index(r.Body, []byte(`:n},queryId:"`))
		requestID = string(r.Body[idStart+13 : idStart+45])
	})

	// c.Visit("https://www.basketball-reference.com/boxscores/")
	c.Visit("https://www.bovada.lv/services/sports/event/v2/events/A/description/football/nfl")

	log.Println(c)
}

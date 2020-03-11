package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/gocolly/colly"
)

func getRefGame(s string, id string) {

	sportMap := map[string]string{
		"nba": "https://www.basketball-reference.com/",
		"nfl": "https://www.pro-football-reference.com/",
		"nhl": "https://www.hockey-reference.com/",
		"mlb": "https://www.baseball-reference.com/",
	}
	// REF_SFX := map[string]string{
	// "nfl": ".htm",
	// "nba": ".html",
	// "mlb": ".shtml",
	// "nhl": "html",
	// }

	// REF_BOX_SFX := map[string]string {
	// 	"nfl": "boxscores/",
	// 	"nba": "boxscores/",
	// 	"mlb": "boxes/",
	// 	"nhl": "boxscores/",
	// }
	rootDomain := sportMap[s]
	fmt.Println(rootDomain)
	fName := "202001260ATL.csv"
	file, _ := os.Create(fName)
	// table id four_factors
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	// writer.Write([]string{"team", "pace", "eFG%", "TOV%", "ORB%", "FT/FTA", "ORtg"})
	writer.Write([]string{"name", "p_id", "pos"})
	c := colly.NewCollector(
		colly.AllowedDomains(rootDomain),
	)

	// Extract product details
	c.OnHTML("#home_starters tbody tr", func(e *colly.HTMLElement) {
		// e.ForEach(".")
		writer.Write([]string{
			e.ChildAttr("a", "title"),
			e.ChildText("span"),
			e.Request.AbsoluteURL(e.ChildAttr("a", "href")),
			"https" + e.ChildAttr("img", "src"),
		})
	})
	c.OnHTML("#vis_starters tbody tr", func(e *colly.HTMLElement) {
		// e.ForEach(".")
		writer.Write([]string{
			e.ChildAttr("a", "title"),
			e.ChildText("span"),
			e.Request.AbsoluteURL(e.ChildAttr("a", "href")),
			"https" + e.ChildAttr("img", "src"),
		})
	})

}

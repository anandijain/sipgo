package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly"
	// "github.com/gocolly/colly/proxy"
)

// Sport : specifying game
type Sport struct {
	Competitions []Competition `json:""`
}

// Competition : specifying game
type Competition struct {
	Events []Event `json:"events"`
	Paths  []Path  `json:"path"`
}

// Event : specifying game
type Event struct {
	id            string         `json:"id"`
	description   string         `json:"description"`
	eventType     string         `json:"type"`
	link          string         `json:"link"`
	status        string         `json:status`
	sport         string         `json:"sport"`
	startTime     int            `json:"startTime"`
	live          bool           `json:"live"`
	awayTeamFirst bool           `json:"awayTeamFirst"`
	denySameGame  string         `json:"denySameGame"`
	teaserAllowed bool           `json:"teaserAllowed"`
	competitionID string         `json:"competitionId"`
	notes         string         `json:"notes"`
	numMarkets    int            `json:"numMarkets"`
	lastModified  int            `json:"lastModified"`
	competitors   []Competitor   `json:"competitors"`
	displayGroups []DisplayGroup `json:"displayGroups"`
}

// DisplayGroup : specifying game
type DisplayGroup struct {
	id            string   `json: id`
	description   string   `json: description`
	defaultType   bool     `json: defaultType`
	alternateType bool     `json: alternateType`
	markets       []Market `json: markets`
	order         int      `json: order`
}

// Competitor : specifying game
type Competitor struct {
	home bool   `json: home`
	id   string `json: id`
	name string `json: name`
}

// Market : specifying game
type Market struct {
	id           string    `json: id`
	description  string    `json: description`
	key          string    `json: key`
	marketTypeID string    `json: marketTypeId`
	status       string    `json: status`
	singleOnly   bool      `json: singleOnly`
	notes        string    `json: notes`
	period       Period    `json: period`
	outcomes     []Outcome `json: outcomes`
}

// Outcome : specifying game
type Outcome struct {
	id           string `json: id`
	description  string `json: description`
	status       string `json: status`
	outcomeType  string `json: outcomeType`
	competitorID string `json: competitorId`
	price        Price  `json: price`
}

// Price : specifying game
type Price struct {
	id         string `json: id`
	handicap   string `json: handicap`
	american   string `json: american`
	decimal    string `json: decimal`
	fractional string `json: fractional`
	malay      string `json: malay`
	indonesian string `json: indonesian`
	hongkong   string `json: hongkong`
}

// Period : specifying game
type Period struct {
	id           string `json: id`
	description  string `json: description`
	abbreviation string `json: abbreviation`
	live         bool   `json: live`
	main         bool   `json: main`
}

// Path : specifying path
type Path struct {
	id          string `json: id`
	link        string `json: link`
	description string `json: description`
	pathType    string `json: pathType`
	sportCode   string `json: sportCode`
	order       int    `json: order`
	leaf        bool   `json: leaf`
	current     bool   `json: current`
}

func main() {
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
		// colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	// c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})

	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	c.OnHTML("p[class=links]", func(e *colly.HTMLElement) {
		fmt.Println(e.Request.AbsoluteURL(e.ChildAttr("a", "href")))
		writer.Write([]string{
			e.Request.AbsoluteURL(e.ChildAttr("a", "href")),
		})
		fmt.Println(e.Request.AbsoluteURL(e.ChildAttr("a", "href")))
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		fmt.Println(e.ChildText("pre"))
	})

	c.OnHTML("html", func(e *colly.HTMLElement) {
		fmt.Println("in herr")
		dat := e.ChildText("body > pre")
		jsonData := dat[strings.Index(dat, "{") : len(dat)-1]
		data := &Sport{}
		err := json.Unmarshal([]byte(jsonData), data)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(data)
	})

	// c.Visit("https://www.basketball-reference.com/boxscores/")
	c.Visit("https://www.bovada.lv/services/sports/event/v2/events/A/description/football/nfl")

	log.Println(c)
}

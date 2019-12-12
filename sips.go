package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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
	ID            string         `json:"id"`
	Description   string         `json:"description"`
	EventType     string         `json:"type"`
	Link          string         `json:"link"`
	Status        string         `json:"status"`
	Sport         string         `json:"sport"`
	StartTime     int            `json:"startTime"`
	Live          bool           `json:"live"`
	AwayTeamFirst bool           `json:"awayTeamFirst"`
	DenySameGame  string         `json:"denySameGame"`
	TeaserAllowed bool           `json:"teaserAllowed"`
	CompetitionID string         `json:"competitionId"`
	Notes         string         `json:"notes"`
	NumMarkets    int            `json:"numMarkets"`
	LastModified  int            `json:"lastModified"`
	Competitors   []Competitor   `json:"competitors"`
	DisplayGroups []DisplayGroup `json:"displayGroups"`
}

// DisplayGroup : specifying game
type DisplayGroup struct {
	ID            string   `json:"id"`
	Description   string   `json:"description"`
	DefaultType   bool     `json:"defaultType"`
	AlternateType bool     `json:"alternateType"`
	Markets       []Market `json:"markets"`
	Order         int      `json:"order"`
}

// Competitor : specifying game
type Competitor struct {
	Home bool   `json:"home"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Market : specifying game
type Market struct {
	ID           string    `json:"id"`
	Description  string    `json:"description"`
	Key          string    `json:"key"`
	MarketTypeID string    `json:"marketTypeId"`
	Status       string    `json:"status"`
	SingleOnly   bool      `json:"singleOnly"`
	Notes        string    `json:"notes"`
	Period       Period    `json:"period"`
	Outcomes     []Outcome `json:"outcomes"`
}

// Outcome : specifying game
type Outcome struct {
	ID           string `json:"id"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	OutcomeType  string `json:"outcomeType"`
	CompetitorID string `json:"competitorId"`
	Price        Price  `json:"price"`
}

// Price : specifying game
type Price struct {
	ID         string `json:"id"`
	Handicap   string `json:"handicap"`
	American   string `json:"american"`
	Decimal    string `json:"decimal"`
	Fractional string `json:"fractional"`
	Malay      string `json:"malay"`
	Indonesian string `json:"indonesian"`
	Hongkong   string `json:"hongkong"`
}

// Period : specifying game
type Period struct {
	ID           string `json:"id"`
	Description  string `json:"description"`
	Abbreviation string `json:"abbreviation"`
	Live         bool   `json:"live"`
	Main         bool   `json:"main"`
}

// Path : specifying path
type Path struct {
	ID          string `json:"id"`
	Link        string `json:"link"`
	Description string `json:"description"`
	PathType    string `json:"pathType"`
	SportCode   string `json:"sportCode"`
	Order       int    `json:"order"`
	Leaf        bool   `json:"leaf"`
	Current     bool   `json:"current"`
}

// ResultStruct : give json to this
type ResultStruct struct {
	Result []map[string]Competition
}

// ResultList : give json to this
type ResultList struct {
	ResultL []Competition
}

func main() {
	res, err := http.Get("https://www.bovada.lv/services/sports/event/v2/events/A/description/football/nfl")
	if err != nil {
		fmt.Println("1")
		log.Fatal(err)
	}

	ret, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		fmt.Println("2")
		log.Fatal(err)
	}

	retString := string(ret)
	// marshErr := json.Unmarshal(ret, t)
	dec := json.NewDecoder(strings.NewReader(retString))
	var c []Competition
	for {
		if err := dec.Decode(&c); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", c[0].Events[0].ID)
		fmt.Printf(string(len(c)))
	}
	json, err := json.MarshalIndent(c, "  ", "    ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(json))

}

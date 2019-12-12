package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Row for CSV headers len 8
type Row struct {
	Sport      string
	GameID     int
	aTeam      string
	hTeam      string
	NumMarkets int
	aML        float64
	hML        float64
	aPS        float64
	hPS        float64
	aHC        float64
	hHC        float64
	gameStart  int
	LastMod    int
}

var sportsMap = map[string]string{
	"nba": "basketball/nba",
	"nfl": "football/nfl",
	"nhl": "hockey/nhl",
}

func makeRow(e Event) Row {
	var r Row
	r.Sport = e.Sport
	gameID, _ := strconv.Atoi(e.ID)
	// if err != nil {
	// 	fmt.Println("gameID weirdness", err)
	// 	gameID := ""
	// }
	r.GameID = gameID

	fmt.Println(e.Competitors)
	if len(e.Competitors) == 2 {
		if e.AwayTeamFirst {
			r.aTeam = e.Competitors[0].Name
			r.hTeam = e.Competitors[1].Name
		} else {
			r.aTeam = e.Competitors[1].Name
			r.hTeam = e.Competitors[0].Name
		}
	} else {
		return r
	}

	// var mkts []Market
	// for _, dg := range e.DisplayGroups {
	// 	for  _, value := range getMainMarkets(dg.Markets) {
	// 	   mkts = append(value, mkts)
	// 	}

	// }
	mkts := e.DisplayGroups[0].Markets
	mainMkts := getMainMarkets(mkts)

	mls := mainMkts["Moneyline"].Outcomes
	if len(mls) == 2 {
		aML, err := strconv.ParseFloat(mls[0].Price.Decimal, 64)
		if err != nil {
			fmt.Println("closed", err)
			r.aML = 0
		}
		hML, err := strconv.ParseFloat(mls[1].Price.Decimal, 64)
		if err != nil {
			fmt.Println("closed", err)
			r.hML = 0
		}

		r.aML = aML
		r.hML = hML

	} else {
		r.aML = 0
		r.hML = 0
	}

	aPS, _ := strconv.ParseFloat(mainMkts["Point Spread"].Outcomes[0].Price.Decimal, 64)
	hPS, _ := strconv.ParseFloat(mainMkts["Point Spread"].Outcomes[1].Price.Decimal, 64)

	aHC, _ := strconv.ParseFloat(mainMkts["Point Spread"].Outcomes[0].Price.Handicap, 64)
	hHC, _ := strconv.ParseFloat(mainMkts["Point Spread"].Outcomes[1].Price.Handicap, 64)

	r.aPS = aPS
	r.hPS = hPS

	r.aHC = aHC
	r.hHC = hHC

	r.LastMod = e.LastModified
	r.gameStart = e.StartTime
	r.NumMarkets = e.NumMarkets

	return r
}

func getMainMarkets(m []Market) map[string]Market {
	ret := make(map[string]Market)
	n := len(m)
	for i := 0; i < n; i++ {
		if m[i].Period.Main {
			ret[m[i].Description] = m[i]
		}
	}
	return ret
}

func setAH(cs [2]Competitor) [2]string {
	if cs[0].Home {
		a := cs[1].Name
		h := cs[0].Name
		return [2]string{a, h}
	}
	h := cs[1].Name
	a := cs[0].Name
	return [2]string{a, h}
}

func toJSON(b []byte) []Competition {
	retString := string(b)
	dec := json.NewDecoder(strings.NewReader(retString))
	var c []Competition
	for {
		if err := dec.Decode(&c); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
	}
	return c
}

func getRows(b []byte) []Row {
	data := toJSON(b)
	var rs []Row
	compEvents := data[0].Events
	numEvents := len(compEvents)

	for i := 0; i < numEvents; i++ {
		r := makeRow(compEvents[i])
		rs = append(rs, r)
	}
	return rs
}

func getSport(s string) []Row {
	res, err := http.Get("https://www.bovada.lv/services/sports/event/v2/events/A/description/" + sportsMap[s])
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

	rs := getRows(ret)
	return rs
}

func main() {

	headers := `{sport,game_id,a_team,h_team,a_ml,h_ml,last_mod,num_markets}`
	fmt.Println(headers)

	rs := getSport("nba")

	for _, row := range rs {
		fmt.Println(row)
	}

}

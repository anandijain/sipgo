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
	aML        float64
	hML        float64
	LastMod    int
	NumMarkets int
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

	if e.AwayTeamFirst {
		r.aTeam = e.Competitors[0].Name
		r.hTeam = e.Competitors[1].Name
	} else {
		r.aTeam = e.Competitors[1].Name
		r.hTeam = e.Competitors[0].Name
	}
	mkts := e.DisplayGroups[0].Markets
	mainMkts := getMainMarkets(mkts)

	aML, _ := strconv.ParseFloat(mainMkts["Moneyline"].Outcomes[0].Price.Decimal, 64)
	hML, _ := strconv.ParseFloat(mainMkts["Moneyline"].Outcomes[1].Price.Decimal, 64)
	r.aML = aML
	r.hML = hML
	r.LastMod = e.LastModified
	r.NumMarkets = e.NumMarkets
	return r
}

func getMainMarkets(m []Market) map[string]Market {
	ret :=  make(map[string]Market)
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

	data := toJSON(ret)

	compEvents := data[0].Events
	numEvents := len(compEvents)
	// var rs []Row

	// for i := 0; i < len(compEvents); i++ {
	// 	game := compEvents[i]
	// 	dgs := game.DisplayGroups
	// 	for j := 0; j < len(dgs); j++ {
	// 		dg := dgs[j]
	// 		mkts := dg.Markets
	// 		for k := 0; k < len(mkts); k++ {
	// 			mkt := mkts[k]
	// 			// fmt.Println(mkt.Description)
	// 			outcomes := mkt.Outcomes
	// 			for l := 0; l < len(outcomes); l++ {
	// 				outcome := outcomes[l]
	// 				fmt.Println(game.ID, game.Description, dg.Description, mkt.Description, outcome.Price.American)
	// 			}
	// 		}
	// 	}
	// }

	for i := 0; i < numEvents; i++ {
		r := makeRow(compEvents[i])
		fmt.Println(r)
	}

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

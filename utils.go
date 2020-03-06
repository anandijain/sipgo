package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

func parseOutcomes(os []Outcome) []float64 {
	var decimals []float64
	for _, o := range os {
		ml, err := strconv.ParseFloat(o.Price.Decimal, 64)
		if err != nil {
			fmt.Println("closed", err)
			ml = 0
		} else {
			decimals = append(decimals, ml)
		}
	}
	return decimals
}

func makeRow(e Event) (Row, bool) {
	var r Row
	var null_row = false
	r.Sport = e.Sport
	gameID, _ := strconv.Atoi(e.ID)
	// if err != nil {
	// 	fmt.Println("gameID weirdness", err)
	// 	gameID := ""
	// }
	r.GameID = gameID

	// fmt.Println(e.Competitors)
	if len(e.Competitors) == 2 {
		if e.AwayTeamFirst {
			r.aTeam = e.Competitors[0].Name
			r.hTeam = e.Competitors[1].Name
		} else {
			r.aTeam = e.Competitors[1].Name
			r.hTeam = e.Competitors[0].Name
		}
	} else {
		null_row = true
	}

	// var mkts []Market
	// for _, dg := range e.DisplayGroups {
	// 	for  _, value := range getMainMarkets(dg.Markets) {
	// 	   mkts = append(value, mkts)
	// 	}

	// }
	mkts := e.DisplayGroups[0].Markets
	mainMkts := getMainMarkets(mkts)
	var mls []Outcome
	if r.Sport == "SOCC" {
		mls = mainMkts["3-Way Moneyline"].Outcomes
	} else {
		mls = mainMkts["Moneyline"].Outcomes
	}
	// spreads := mainMkts["Point Spread"].Outcomes
	parsedMLs := parseOutcomes(mls)
	// parsedSpreads := parseOutcomes(spreads)
	if r.Sport == "SOCC" {
		if len(parsedMLs) == 3 {
			r.aML = parsedMLs[0]
			r.hML = parsedMLs[1]
			r.drawML = parsedMLs[2]
			} else {
				return r, null_row
		}
	} else {
		if len(parsedMLs) == 2 {
			r.aML = parsedMLs[0]
			r.hML = parsedMLs[1]
			r.drawML = 0.
			} else {
				return r, null_row
		}
	}
	// if len(parsedSpreads) != 2 {
	// 	return r, null_row
	// } else {
	// 	r.aPS = parsedSpreads[0]
	// 	r.hPS = parsedSpreads[1]
	// }

	// aPS, _ := strconv.ParseFloat(spreads[0].Price.Decimal, 64)
	// hPS, _ := strconv.ParseFloat(spreads[1].Price.Decimal, 64)

	// aHC, _ := strconv.ParseFloat(spreads[0].Price.Handicap, 64)
	// hHC, _ := strconv.ParseFloat(spreads[1].Price.Handicap, 64)

	// r.aPS = aPS
	// r.hPS = hPS

	// r.aHC = aHC
	// r.hHC = hHC

	r.LastMod = e.LastModified
	r.gameStart = e.StartTime
	r.NumMarkets = e.NumMarkets
	if r.aTeam == "" {
		fmt.Println("null row")
		null_row = true
	}
	return r, null_row
}

func makeScore(s Score) shortScore {
	var r shortScore
	r.Period = s.Clock.PeriodNumber
	r.Seconds = s.Clock.RelativeGameTimeInSecs
	r.IsTicking = s.Clock.IsTicking
	// r.NumberOfPeriods = s.Clock.NumberOfPeriods
	if len(s.Competitors) != 2 {
		return r
	}
	
	if s.Competitors[0].Name == "" {
		fmt.Println("broke")
	}
	r.aTeam = s.Competitors[0].Name
	r.hTeam = s.Competitors[1].Name
	r.aPts = s.LatestScore.Visitor
	r.hPts = s.LatestScore.Home
	r.Status = s.GameStatus
	r.lastMod = s.LastUpdated
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
			fmt.Println("json couldnt -> []Competition")
			log.Fatal(err)
		}
	}
	return c
}

func scoreToJSON(b []byte) Score {
	retString := string(b)
	dec := json.NewDecoder(strings.NewReader(retString))
	var s Score
	if err := dec.Decode(&s); err == io.EOF {
		fmt.Println("json couldnt -> Score")
		log.Fatal(err)
	} else if err != nil {
		fmt.Println("json couldnt -> Score")
		log.Fatal(err)
	}
	return s
}

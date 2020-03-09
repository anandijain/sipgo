package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
)

var pgSchema = `CREATE TABLE rows(
	id serial NOT NULL PRIMARY KEY, 
	game_id int NOT NULL, 
	sport character(4), 
	league character(128), 
	comp character(128), 
	country character(128), 
	region character(128), 
	a_team character(64), 
	h_team character(64), 
	num_markets int, 
	a_ml real, 
	h_ml real, 
	draw_ml real,  
	game_start timestamp, 
	last_mod timestamp,
	period int,
	secs int, 
	is_ticking boolean,
	a_pts int,
	h_pts int,
	status character(32)`

func makeLine(e Event) (Line, bool) {
	var r Line
	var null_row = false
	r.Sport = e.Sport
	gameID, _ := strconv.Atoi(e.ID)
	r.GameID = gameID

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
	mkts := e.DisplayGroups[0].Markets
	mainMkts := getMainMarkets(mkts)
	var mls []Outcome
	if includes(drawSports, r.Sport) {
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

	r.LastMod = e.LastModified
	r.gameStart = e.StartTime
	r.NumMarkets = e.NumMarkets
	if r.aTeam == "" {
		null_row = true
	}
	return r, null_row
}

func getScores(ids []int, concurrencyLimit int) map[int]shortScore {
	var results []concurrentRes
	scores := make(map[int]shortScore)
	semaphoreChan := make(chan struct{}, concurrencyLimit)
	resultsChan := make(chan *concurrentRes)

	// make sure we close these channels when we're done with them
	defer func() {
		close(semaphoreChan)
		close(resultsChan)
	}()

	for i, g_id := range ids {
		go func(i int, g_id int) {
			semaphoreChan <- struct{}{}
			s, err := getScore(strconv.Itoa(g_id))
			result := concurrentRes{i, s, err}
			resultsChan <- &result
			<-semaphoreChan
		}(i, g_id)
	}
	for {
		result := <-resultsChan
		results = append(results, *result)
		if len(results) == len(ids) {
			break
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].index < results[j].index
	})
	for _, res := range results {
		scores[res.res.GameID] = res.res
	}
	return scores
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

func getScore(s string) (shortScore, error) {
	ret, err := req(scoreRoot + s)
	if err != nil {
		fmt.Println("2")
		log.Fatal(err)
	}
	data := scoreFromBytes(ret)
	r := makeScore(data)
	return r, err
}

func lineLooperz(s string) {
	_, w := initCSV("lines.csv", lineHeaders)
	for {
		lines, _ := getLines(s)
		to_write := linesToCSV(lines)
		w.WriteAll(to_write)
	}
}

func parseLines(b []byte) map[int]Line {
	data := toJSON(b)
	rs := make(map[int]Line)
	// var events []Event
	for _, ev := range data {
		es := ev.Events
		for _, e := range es {
			r, null_row := makeLine(e)
			if null_row == false {
				rs[r.GameID] = r
			}
		}
	}

	return rs
}

func scoreToCSV(s shortScore) []string {
	ret := []string{
		fmt.Sprint(s.GameID),
		s.aTeam,
		s.hTeam,
		fmt.Sprint(s.Period),
		fmt.Sprint(s.Seconds),
		fmt.Sprint(s.IsTicking),
		fmt.Sprint(s.aPts),
		fmt.Sprint(s.hPts),
		s.Status,
		fmt.Sprint(s.lastMod),
	}
	return ret
}

func scoresToCSV(data map[int]shortScore) [][]string {
	var recs [][]string
	for _, r := range data {
		row := scoreToCSV(r)
		recs = append(recs, row)
	}
	return recs
}

func getLines(s string) (map[int]Line, error) {
	ret, err := req(s)
	rs := parseLines(ret)
	return rs, err
}
func lineToCSV(r Line) []string {
	ret := []string{r.Sport, fmt.Sprint(r.GameID), r.aTeam, r.hTeam, fmt.Sprint(r.NumMarkets), fmt.Sprint(r.aML), fmt.Sprint(r.hML), fmt.Sprint(r.drawML),
		fmt.Sprint(r.gameStart), fmt.Sprint(r.LastMod)}
	return ret
}

func linesToCSV(data map[int]Line) [][]string {
	var recs [][]string
	for _, r := range data {
		row := lineToCSV(r)
		recs = append(recs, row)
	}
	return recs
}
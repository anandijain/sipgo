package main

import (
	"fmt"
	"sort"
	"strconv"
)

func getScoresConcurrent(ids []int, concurrencyLimit int) map[int]shortScore {
	var results []concurrentResScore
	scores := make(map[int]shortScore)
	semaphoreChan := make(chan struct{}, concurrencyLimit)
	resultsChan := make(chan *concurrentResScore)

	// make sure we close these channels when we're done with them
	defer func() {
		close(semaphoreChan)
		close(resultsChan)
	}()

	for i, id := range ids {
		go func(i int, id int) {
			semaphoreChan <- struct{}{}
			s, err := getScore(strconv.Itoa(id))
			result := concurrentResScore{i, s, err}
			resultsChan <- &result
			<-semaphoreChan
		}(i, id)
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
 
func makeScore(s Score) (shortScore, bool) {
	var r shortScore
	nullRow := false
	fmt.Println(s.EventID)
	r.GameID = s.EventID
	r.Period = s.Clock.PeriodNumber
	r.Seconds = s.Clock.RelativeGameTimeInSecs
	r.IsTicking = s.Clock.IsTicking
	// r.NumberOfPeriods = s.Clock.NumberOfPeriods
	if len(s.Competitors) != 2 {
		nullRow = true
	}
	
	if s.Competitors[0].Name == "" {
		nullRow = true
	}
	r.aTeam = s.Competitors[0].Name
	r.hTeam = s.Competitors[1].Name
	r.aPts = s.LatestScore.Visitor
	r.hPts = s.LatestScore.Home
	r.Status = s.GameStatus
	r.lastMod = s.LastUpdated
	return r, nullRow
}

func getScore(s string) (shortScore, error) {
	ret, err := req(scoreRoot + s)
	if err != nil {
		fmt.Println("get score err", err)
	}
	data := scoreFromBytes(ret)
	r, _ := makeScore(data)
	return r, err
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
func setCompetitorsLine(e Event, l Line) (Line, bool) {
	var nullRow = false
	if len(e.Competitors) == 2 {
		if e.AwayTeamFirst {
			l.aTeam = e.Competitors[0].Name
			l.hTeam = e.Competitors[1].Name
		} else {
			l.aTeam = e.Competitors[1].Name
			l.hTeam = e.Competitors[0].Name
		}
	} else {
		nullRow = true
	}
	return l, nullRow
}
func parseMarketsToLine(e Event, l Line) (Line, bool) {
	var nullRow = false
	mkts := e.DisplayGroups[0].Markets
	mainMkts := getMainMarkets(mkts)
	var mls []Outcome
	switch {
	case includes(drawSports, l.Sport):
		mls = mainMkts["3-Way Moneyline"].Outcomes
	case l.Sport == "BOXI":
		mls = mainMkts["To Win the Bout"].Outcomes
	case l.Sport == "MMA":
		mls = mainMkts["Fight Winner"].Outcomes
	case l.Sport == "DART":
		mls = mainMkts["Winner"].Outcomes
	default:
		mls = mainMkts["Moneyline"].Outcomes
	}
	parsedMLs := parseOutcomes(mls)
	l.aML = parsedMLs[0]
	l.hML = parsedMLs[1]
	if len(parsedMLs) == 3 {
		l.drawML = parsedMLs[2]
	}
	return l, nullRow
}


func makeLine(e Event) (Line, bool) {
	var l Line
	var nullRow = false
	l.Sport = e.Sport
	gameID, _ := strconv.Atoi(e.ID)
	l.GameID = gameID
	l, nullRow = setCompetitorsLine(e, l)
	l, nullRow = parseMarketsToLine(e, l)
	l.LastMod = e.LastModified
	l.gameStart = e.StartTime
	l.NumMarkets = e.NumMarkets
	if l.aTeam == "" {
		nullRow = true
	}
	return l, nullRow
}
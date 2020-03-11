package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

var concurrencyLim = 10

// len 19
var supportedSports = []string{"soccer", "esports", "ufc-mma", "football",
	"boxing", "basketball", "tennis", "rugby-league",
	"table-tennis", "hockey", "volleyball", "rugby-union",
	"darts", "aussie-rules", "badminton", "baseball",
	"handball", "cricket", "snooker"}

var lineRoot = "https://www.bovada.lv/services/sports/event/v2/events/A/description/"
var scoreRoot = "https://www.bovada.lv/services/sports/results/api/v1/scores/"
var drawSports = []string{"SOCC", "RUGU", "RUGL"}
var lineHeaders = []string{"sport", "game_id", "a_team", "h_team", "num_markets", "a_ml", "h_ml", "draw_ml", "last_mod"}
var scoreHeaders = []string{"game_id", "a_team", "h_team", "period", "secs", "is_ticking", "a_pts", "h_pts", "status", "last_mod_score"}
var allHeaders = []string{"game_id", "sport", "league", "comp", "country", "region", "a_team", "h_team", "num_markets", "a_ml",
	"h_ml", "draw_ml", "game_start", "last_mod", "period", "secs", "is_ticking", "a_pts",
	"h_pts", "status"}

func parseOutcomes(os []Outcome) map[int]float64 {
	decimals := make(map[int]float64)
	for i, o := range os {
		ml, err := strconv.ParseFloat(o.Price.Decimal, 64)
		if err != nil {
			// fmt.Println("closed", err)
			ml = 0
		} else {
			decimals[i] = ml
		}
	}
	return decimals
}

func indexStrs(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

func includes(vs []string, t string) bool {
	return indexStrs(vs, t) >= 0
}

func setCompetitors(e Event, r Row) (Row, bool) {
	var nullRow = false
	if len(e.Competitors) == 2 {
		if e.AwayTeamFirst {
			r.aTeam = e.Competitors[0].Name
			r.hTeam = e.Competitors[1].Name
		} else {
			r.aTeam = e.Competitors[1].Name
			r.hTeam = e.Competitors[0].Name
		}
	} else {
		nullRow = true
	}
	return r, nullRow
}

func parseMarkets(e Event, r Row) (Row, bool) {
	var nullRow = false
	mkts := e.DisplayGroups[0].Markets
	mainMkts := getMainMarkets(mkts)
	var mls []Outcome
	switch {
	case includes(drawSports, r.Sport):
		mls = mainMkts["3-Way Moneyline"].Outcomes
	case r.Sport == "BOXI":
		mls = mainMkts["To Win the Bout"].Outcomes
	case r.Sport == "MMA":
		mls = mainMkts["Fight Winner"].Outcomes
	case r.Sport == "DART":
		mls = mainMkts["Winner"].Outcomes
	default:
		mls = mainMkts["Moneyline"].Outcomes
	}
	parsedMLs := parseOutcomes(mls)
	r.aML = parsedMLs[0]
	r.hML = parsedMLs[1]
	r.drawML = parsedMLs[2]
	return r, nullRow
}

func makeLineToRow(e Event) (Row, bool) {
	var r Row
	var nullRow = false
	r.Sport = e.Sport
	gameID, _ := strconv.Atoi(e.ID)
	r.GameID = gameID
	r, nullRow = setCompetitors(e, r)
	r, nullRow = parseMarkets(e, r)
	r.LastMod = e.LastModified
	r.gameStart = e.StartTime
	r.NumMarkets = e.NumMarkets
	if r.aTeam == "" {
		nullRow = true
	}
	return r, nullRow
}

func getMainMarkets(ms []Market) map[string]Market {
	ret := make(map[string]Market)
	for _, m := range ms {
		if m.Period.Main {
			ret[m.Description] = m
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
			print(b)
			// log.Fatal(err)
		}
	}
	return c
}

func scoreFromBytes(b []byte) Score {
	retString := string(b)
	dec := json.NewDecoder(strings.NewReader(retString))
	var s Score
	if err := dec.Decode(&s); err == io.EOF {

		fmt.Println("json couldnt -> Score", err)
	} else if err != nil {
		fmt.Println("json couldnt -> Score", err)
	}
	return s
}

func scoresFromBytes(b []byte) Scores {
	retString := string(b)
	dec := json.NewDecoder(strings.NewReader(retString))
	var s Scores
	for {
		if err := dec.Decode(&s); err == io.EOF {
			break
			fmt.Println("json couldnt -> Score", err)
		} else if err != nil {
			fmt.Println("json couldnt -> Score", err)
			print(b)
		}
	}
	print(s)
	return s
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func initCSV(fn string, header []string) (*os.File, *csv.Writer) {
	f, err := os.Create(fn)
	checkError("Cannot create file", err)
	// defer f.Close()

	w := csv.NewWriter(f)
	w.Write(header)
	// defer w.Flush()
	return f, w
}

func addScores(rs map[int]Row, concurrencyLimit int) map[int]Row {
	var results []concurrentResRow
	scores := make(map[int]Row)
	semaphoreChan := make(chan struct{}, concurrencyLimit)
	resultsChan := make(chan *concurrentResRow)

	defer func() {
		close(semaphoreChan)
		close(resultsChan)
	}()
	var err error
	i := 0
	for _, r := range rs {
		go func(i int, r Row) {
			semaphoreChan <- struct{}{}
			r, err = addScore(r)
			result := concurrentResRow{i, r, err}
			i++
			resultsChan <- &result
			<-semaphoreChan
		}(i, r)
	}
	for {
		result := <-resultsChan
		results = append(results, *result)
		if len(results) == len(rs) {
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

func addScore(r Row) (Row, error) {
	ret, err := req(scoreRoot + strconv.Itoa(r.GameID))
	if err != nil {
		fmt.Println("2")
		log.Fatal(err)
	}
	s := scoreFromBytes(ret)

	r.Period = s.Clock.PeriodNumber
	r.Seconds = s.Clock.RelativeGameTimeInSecs
	r.IsTicking = s.Clock.IsTicking
	if len(s.Competitors) != 2 {
		return r, err
	}

	if s.Competitors[0].Name == "" {
		fmt.Println("broke")
	}

	r.aPts, _ = strconv.Atoi(s.LatestScore.Visitor)
	r.hPts, _ = strconv.Atoi(s.LatestScore.Home)
	r.Status = s.GameStatus
	r.lastMod = s.LastUpdated
	return r, err
}

// league country competition/game sport
func rowToCSV(r Row) []string {
	ret := []string{
		fmt.Sprint(r.GameID),
		r.Sport,
		r.League,
		r.Comp,
		r.Country,
		r.Region,
		r.aTeam,
		r.hTeam,
		fmt.Sprint(r.NumMarkets),
		fmt.Sprint(r.aML),
		fmt.Sprint(r.hML),
		fmt.Sprint(r.drawML),
		fmt.Sprint(r.gameStart),
		fmt.Sprint(r.LastMod),
		fmt.Sprint(r.Period),
		fmt.Sprint(r.Seconds),
		fmt.Sprint(r.IsTicking),
		fmt.Sprint(r.aPts),
		fmt.Sprint(r.hPts),
		r.Status,
		fmt.Sprint(r.lastMod)}
	return ret
}

func compRows(prev map[int]Row, cur map[int]Row) map[int]Row {
	toWrite := make(map[int]Row)
	for id, r := range cur {
		if !reflect.DeepEqual(r, prev[id]) {
			toWrite[r.GameID] = r
		}
	}
	return toWrite
}

func rowsToCSVFormat(data map[int]Row) [][]string {
	var recs [][]string
	for _, r := range data {
		row := rowToCSV(r)
		recs = append(recs, row)
	}
	return recs
}

func loopN(s string, n int, fn string) {
	_, rowWriter := initCSV(fn, allHeaders)

	for i := 1; i <= n; i++ {
		rs := getRows(lineRoot + s)
		rowsToWrite := rowsToCSVFormat(rs)
		rowWriter.WriteAll(rowsToWrite)
	}
}

func idsFromLines(rs map[int]Line) []int {
	var ids []int
	for k := range rs {
		ids = append(ids, k)
	}
	return ids
}

func parsePaths(ps []Path) Row {
	var categories Row
	for _, p := range ps {
		switch t := p.PathType; t {
		case "COUNTRY":
			categories.Country = strings.Replace(p.Description, ",", "", -1)
		case "REGION":
			categories.Region = strings.Replace(p.Description, ",", "", -1)
		case "LEAGUE":
			categories.League = strings.Replace(p.Description, ",", "", -1)
		case "COMPETITION":
			categories.Comp = strings.Replace(p.Description, ",", "", -1)
		}
	}
	return categories
}

func parseLines(b []byte) map[int]Line {
	data := toJSON(b)
	rs := make(map[int]Line)
	// var events []Event
	for _, ev := range data {
		es := ev.Events
		for _, e := range es {
			r, nullRow := makeLine(e)
			if nullRow == false {
				rs[r.GameID] = r
			}
		}
	}

	return rs
}

func getRowsNoScore(s string) (map[int]Row, error) {
	ret, err := req(lineRoot + s)
	rs := parseLinesToRows(ret)
	return rs, err
}

func parallelGetRows(sports []string, concurrencyLimit int) map[int]Row {
	results := make(map[int]Row)
	semaphoreChan := make(chan struct{}, concurrencyLimit)
	resultsChan := make(chan *concurrentMapRowResult)
	defer func() {
		close(semaphoreChan)
		close(resultsChan)
	}()
	for i, s := range sports {
		go func(i int, s string) {
			semaphoreChan <- struct{}{}
			// rs := getRows(s)
			rs, _ := getRowsNoScore(s)
			result := &concurrentMapRowResult{i, s, rs}
			resultsChan <- result
			<-semaphoreChan

		}(i, s)
	}
	for {
		result := <-resultsChan
		for _, res := range result.res {
			results[res.GameID] = res
		}
		if len(results) == len(sports) {
			break
		}
	}

	return results
}

func separateConcurrent(sports []string, concurrencyLimit int) map[int]Row {
	rs := parallelGetRows(sports, concurrencyLimit)
	rs = addScores(rs, concurrencyLim)
	return rs
}

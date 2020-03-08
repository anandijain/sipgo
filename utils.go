package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var lineRoot = "https://www.bovada.lv/services/sports/event/v2/events/A/description/"
var scoreRoot = "https://www.bovada.lv/services/sports/results/api/v1/scores/"
var drawSports = []string{"SOCC", "RUGU", "RUGL"}
var lineHeaders = []string{"sport", "game_id", "a_team", "h_team", "num_markets", "a_ml", "h_ml", "draw_ml", "last_mod"}
var scoreHeaders = []string{"game_id", "a_team", "h_team", "period", "secs", "is_ticking", "a_pts", "h_pts", "status", "last_mod_score"}
var allHeaders = []string{"game_id", "sport", "league", "comp", "country", "region", "a_team", "h_team", "num_markets", "a_ml",
	"h_ml", "draw_ml", "game_start", "last_mod", "period", "secs", "is_ticking", "a_pts",
	"h_pts", "status"}

var schema = `CREATE TABLE rows(
	id int NOT NULL AUTO_INCREMENT, 
	game_id int, 
	sport varchar(4), 
	league varchar(128), 
	comp varchar(128), 
	country varchar(128), 
	region varchar(128), 
	a_team varchar(64), 
	h_team varchar(64), 
	num_markets int, 
	a_ml float, 
	h_ml float, 
	draw_ml float,  
	game_start int, 
	last_mod int,
	period int,
	secs int, 
	is_ticking boolean,
	a_pts int,
	h_pts int,
	status varchar(32)
	PRIMARY KEY (id))`

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

func index_strs(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}
func includes(vs []string, t string) bool {
	return index_strs(vs, t) >= 0
}

func makeLineToRow(e Event) (Row, bool) {
	var r Row
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

func scoreFromBytes(b []byte) Score {
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

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
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
	r.aPts = s.LatestScore.Visitor
	r.hPts = s.LatestScore.Home
	r.Status = s.GameStatus
	r.lastMod = s.LastUpdated
	return r, err
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

func getLines(s string) (map[int]Line, error) {
	ret, err := req(s)
	rs := parseLines(ret)
	return rs, err
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
	to_write := make(map[int]Row)
	for cur_id, r := range cur {
		if !reflect.DeepEqual(r, prev[cur_id]) {
			to_write[r.GameID] = r
		}
	}
	return to_write
}

func rowsToCSVFormat(data map[int]Row) [][]string {
	var recs [][]string
	for _, r := range data {
		row := rowToCSV(r)
		recs = append(recs, row)
	}
	return recs
}


func loop_n(s string, n int, fn string) {
	_, rowWriter := initCSV(fn, allHeaders)

	for i := 1; i <= n; i++ {
		rs := grabRows(lineRoot + s)
		rowsToWrite := rowsToCSVFormat(rs)
		rowWriter.WriteAll(rowsToWrite)
	}
}
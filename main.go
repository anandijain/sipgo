// credit montanaflynn/pget.go https://gist.github.com/montanaflynn/ea4b92ed640f790c4b9cee36046a5383
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"time"
)

var lineHeaders = []string{"sport", "game_id", "a_team", "h_team", "num_markets", "a_ml", "h_ml", "draw_ml", "last_mod"}
var scoreHeaders = []string{"game_id", "a_team", "h_team", "period", "secs", "is_ticking", "a_pts", "h_pts", "status", "last_mod_score"}
var allHeaders = []string{"sport", "game_id", "a_team", "h_team", "num_markets", "a_ml", "h_ml", "draw_ml", "last_mod", "period", "secs", "is_ticking", "a_pts", "h_pts", "status", "last_mod_score"}

var lineRoot = "https://www.bovada.lv/services/sports/event/v2/events/A/description/"
var scoreRoot = "https://www.bovada.lv/services/sports/results/api/v1/scores/"

type concurrentRes struct {
	index int
	res   shortScore
	err   error
}

type concurrentResRow struct {
	index int
	res   Row
	err   error
}

// Line for CSV headers len 10
type Line struct {
	Sport      string
	GameID     int
	aTeam      string
	hTeam      string
	NumMarkets int
	aML        float64
	hML        float64
	drawML     float64
	gameStart  int
	LastMod    int
}

// score for CSV len 10
type shortScore struct {
	GameID    int
	aTeam     string
	hTeam     string
	Period    int
	Seconds   int
	IsTicking bool
	aPts      string
	hPts      string
	Status    string
	lastMod   string
}

// Line for CSV headers len 17
type Row struct {
	Sport      string
	GameID     int
	aTeam      string
	hTeam      string
	NumMarkets int
	aML        float64
	hML        float64
	drawML     float64
	gameStart  int
	LastMod    int
	Period     int
	Seconds    int
	IsTicking  bool
	aPts       string
	hPts       string
	Status     string
	lastMod    string
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

func rowToCSV(r Row) []string {
	ret := []string{
		r.Sport,
		fmt.Sprint(r.GameID),
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

func rowsToCSV(data map[int]Row) [][]string {
	var recs [][]string
	for _, r := range data {
		row := rowToCSV(r)
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

func getLinesForRows(s string) (map[int]Row, error) {
	ret, err := req(s)
	rs := parseLinesToRows(ret)
	return rs, err
}

func parseLinesToRows(b []byte) map[int]Row {
	data := toJSON(b)
	rs := make(map[int]Row)
	// var events []Event
	for _, ev := range data {
		es := ev.Events
		for _, e := range es {
			r, null_row := makeLineToRow(e)
			if null_row == false {
				rs[r.GameID] = r
			}
		}
	}

	return rs
}

func req(s string) ([]byte, error) {
	res, httperr := http.Get(s)
	if httperr != nil {
		fmt.Println("1")
		log.Fatal(httperr)
	}
	ret, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return ret, httperr
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

func idsFromLines(rs map[int]Line) []int {
	var ids []int
	for k := range rs {
		ids = append(ids, k)
	}
	return ids
}

func addScores(rs map[int]Row, concurrencyLimit int) map[int]Row {
	var results []concurrentResRow
	scores := make(map[int]Row)
	semaphoreChan := make(chan struct{}, concurrencyLimit)
	resultsChan := make(chan *concurrentResRow)

	// make sure we close these channels when we're done with them
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
	r.aPts = s.LatestScore.Visitor
	r.hPts = s.LatestScore.Home
	r.Status = s.GameStatus
	r.lastMod = s.LastUpdated
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

func grabRows(s string) map[int]Row {
	rs, _ := getLinesForRows(s)
	rs = addScores(rs, 16)
	return rs
}

func comp(prev map[int]Row, cur map[int]Row) map[int]Row {
	to_write := make(map[int]Row)
	for cur_id, r := range cur {
		if !reflect.DeepEqual(r, prev[cur_id]) {
			to_write[r.GameID] = r
		}
	}
	return to_write
}

func looperz(s string, fn string) {

	_, w := initCSV(fn, allHeaders)
	url := lineRoot + s
	prev := grabRows(url)
	initWrite := rowsToCSV(prev)
	w.WriteAll(initWrite)
	cur := grabRows(url)

	for {
		diff := comp(prev, cur)
		rowsToWrite := rowsToCSV(diff)
		fmt.Println(len(rowsToWrite), "# of changes", time.Now())
		// for _, r := range rowsToWrite {
		// 	fmt.Println(r)
		// }
		prev = cur
		cur = grabRows(url)

		w.WriteAll(rowsToWrite)
	}
}

func loop_n(s string, n int, fn string) {
	_, rowWriter := initCSV(fn, allHeaders)

	for i := 1; i <= n; i++ {
		rs := grabRows(lineRoot + s)
		rowsToWrite := rowsToCSV(rs)
		rowWriter.WriteAll(rowsToWrite)
	}
}

func main() {
	looperz("", "rows.csv")
}

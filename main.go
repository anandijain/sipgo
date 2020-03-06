// credit montanaflynn/pget.go https://gist.github.com/montanaflynn/ea4b92ed640f790c4b9cee36046a5383
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
	// "google.golang.org/appengine"
)

type concurrentRes struct {
	index int
	res   shortScore
	err   error
}

// Row for CSV headers len 8
type Row struct {
	Sport      string
	GameID     int
	aTeam      string
	hTeam      string
	NumMarkets int
	aML        float64
	hML        float64
	drawML     float64
	// aPS        float64
	// hPS        float64
	// aHC        float64
	// hHC        float64
	gameStart int
	LastMod   int
}

type shortScore struct {
	GameID    int
	aTeam     string
	hTeam     string
	Period    int
	Seconds   int
	IsTicking bool
	// NumberOfPeriods int
	aPts    string
	hPts    string
	Status  string
	lastMod string
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

func rowToCSV(r Row) []string {
	ret := []string{r.Sport, fmt.Sprint(r.GameID), r.aTeam, r.hTeam, fmt.Sprint(r.NumMarkets), fmt.Sprint(r.aML), fmt.Sprint(r.hML), fmt.Sprint(r.drawML),
		fmt.Sprint(r.gameStart), fmt.Sprint(r.LastMod)}
	return ret
}

func rowsToCSV(data []Row) [][]string {
	var recs [][]string
	for _, r := range data {
		row := rowToCSV(r)
		recs = append(recs, row)
	}
	return recs
}

func scoresToCSV(data []shortScore) [][]string {
	var recs [][]string
	for _, r := range data {
		row := scoreToCSV(r)
		recs = append(recs, row)
	}
	return recs
}

var sportsMap = map[string]string{
	"nba":  "basketball/nba",
	"nfl":  "football/nfl",
	"nhl":  "hockey/nhl",
	"mlb":  "baseball/mlb",
	"FOOT": "football/",
	"BASK": "basketball/",
}

func getLines(s string) []Row {
	res, err := http.Get("https://www.bovada.lv/services/sports/event/v2/events/A/description/" + s)
	if err != nil {
		log.Fatal(err)
	}
	ret, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		log.Fatal(err)
	}

	rs := getRows(ret)
	return rs
}

func getRows(b []byte) []Row {
	data := toJSON(b)
	var rs []Row
	// var events []Event
	drawSports := []string{"SOCC", "RUGU", "RUGL"}
	for _, ev := range data {
		es := ev.Events
		for _, e := range es {
			r, null_row := makeRow(e, drawSports)
			if null_row == false {
				rs = append(rs, r)
			}
		}
	}

	return rs
}

func getScore(s string) (shortScore, error) {
	url := "https://www.bovada.lv/services/sports/results/api/v1/scores/" + s
	res, httperr := http.Get(url)
	// fmt.Println(res)
	if httperr != nil {
		fmt.Println("1")
		log.Fatal(httperr)
	}
	ret, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		fmt.Println("2")
		log.Fatal(err)
	}
	gameID, _ := strconv.Atoi(s)
	data := scoreToJSON(ret)
	r := makeScore(data)
	r.GameID = gameID
	return r, httperr
}

func idsFromRows(rs []Row) []int {
	var ids []int
	for _, r := range rs {
		ids = append(ids, r.GameID)
	}
	return ids
}

func getScores(ids []int, concurrencyLimit int) []shortScore {
	var scores []shortScore
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
	var results []concurrentRes

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
		scores = append(scores, res.res)
	}
	return scores
}

func sportWithScores(s string) ([]Row, []shortScore) {

	nba := getLines(s)
	ids := idsFromRows(nba)
	scores := getScores(ids, 2)

	return nba, scores
}

func lineLooperz(s string) {
	headers := []string{"sport", "game_id", "a_team", "h_team", "num_markets", "a_ml", "h_ml", "draw_ml", "last_mod"}
	// _, w := initCSV("lines.csv", headers)
	_, w := initCSV("lines.csv", headers)
	w.Flush()
	// f.Close()
	prev := time.Now()
	now := time.Now()
	delta := now.Sub(prev)
	i := 0
	for true {
		lines := getLines(s)
		to_write := rowsToCSV(lines)
		// fmt.Println(to_write)
		w.WriteAll(to_write)

		i = i + 1
		delta = now.Sub(prev)
		prev = now
		now = time.Now()
		fmt.Println("%s", i, delta)
		time.Sleep(time.Duration(10) * time.Second)
	}

}
func scoreLooperz(s string, fn string) {
	scoreHeaders := []string{"game_id", "a_team", "h_team", "period", "secs", "is_ticking", "a_pts", "h_pts", "status", "last_mod"}
	_, w := initCSV(fn, scoreHeaders)
	w.Flush()
	// f.Close()
	prev := time.Now()
	now := time.Now()
	delta := now.Sub(prev)
	i := 0
	for true {
		w.WriteAll(getScoreRows(s))
		i = i + 1
		delta = now.Sub(prev)
		prev = now
		now = time.Now()
		fmt.Println("%s", i, delta)
		time.Sleep(time.Duration(10) * time.Second)
	}

}
func getScoreRows(s string) [][]string {
	lines := getLines(s)
	ids := idsFromRows(lines)
	scores := getScores(ids, 2)
	to_write := scoresToCSV(scores)
	return to_write
}

func main() {
	start := time.Now()

	headers := []string{"sport", "game_id", "a_team", "h_team", "num_markets", "a_ml", "h_ml", "draw_ml", "last_mod"}
	scoreHeaders := []string{"game_id", "a_team", "h_team", "period", "secs", "is_ticking", "a_pts", "h_pts", "status", "last_mod"}

	_, lineWriter := initCSV("lines2.csv", headers)
	_, scoreWriter := initCSV("scores2.csv", scoreHeaders)
	// lineLooperz("")
	rs := getLines("")
	rowsToWrite := rowsToCSV(rs)
	ids := idsFromRows(rs)
	scs := getScores(ids, len(ids))
	scoresToWrite := scoresToCSV(scs)

	lineWriter.WriteAll(rowsToWrite)
	scoreWriter.WriteAll(scoresToWrite)

	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Println(scs)
	fmt.Println(elapsed)
}

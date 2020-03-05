package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	// "google.golang.org/appengine"
)

// var db *sql.DB

func rowToCSV(r Row) []string {
	ret := []string{r.Sport, fmt.Sprint(r.GameID), r.aTeam, r.hTeam, fmt.Sprint(r.NumMarkets), fmt.Sprint(r.aML), fmt.Sprint(r.hML),
		fmt.Sprint(r.gameStart), fmt.Sprint(r.LastMod)}
	return ret
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

	for _, ev := range data {
		es := ev.Events
		for _, e := range es {
			r, null_row := makeRow(e)
			if null_row == false {
				rs = append(rs, r)
			}
		}
	}

	return rs
}

func getScore(s string) shortScore {
	url := "https://www.bovada.lv/services/sports/results/api/v1/scores/" + s
	res, err := http.Get(url)
	// fmt.Println(res)
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
	gameID, _ := strconv.Atoi(s)
	data := scoreToJSON(ret)
	r := makeScore(data)
	r.GameID = gameID
	return r
}

func idsFromRows(rs []Row) []int {
	var ids []int
	for _, r := range rs {
		ids = append(ids, r.GameID)
	}
	return ids
}

func getScores(ids []int) []shortScore {
	var scores []shortScore
	for _, id := range ids {
		game_id := strconv.Itoa(id)
		s := getScore(game_id)
		scores = append(scores, s)
	}
	return scores
}

func sportWithScores(s string) ([]Row, []shortScore) {

	nba := getLines(s)
	ids := idsFromRows(nba)
	scores := getScores(ids)

	return nba, scores
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func toCSV(w csv.Writer, data [][]string) {

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

func lineLooperz(s string) {
	headers := []string{"sport", "game_id", "a_team", "h_team", "a_ml", "h_ml", "last_mod", "num_markets"}
	// _, w := initCSV("lines.csv", headers)
	_, w := initCSV("lines.csv", headers)
	w.Flush()
	// f.Close()
	prev := time.Now()
	now := time.Now()
	i := 0
	for true {
		lines := getLines(s)
		to_write := rowsToCSV(lines)
		fmt.Println(to_write)
		w.WriteAll(to_write)
		
		i = i + 1
		now  time.Now()
		delta
		fmt.Println("%s", i)
		time.Sleep(time.Duration(10)*time.Second)
	}

}
func writeScores(s string){
	scoreHeaders := []string{"game_id", "a_team", "h_team", "period", "secs", "is_ticking", "a_pts", "h_pts", "status", "last_mod"}
	_, w := initCSV("scores.csv", scoreHeaders)
	
	lines := getLines(s)
	ids := idsFromRows(lines)
	scores := getScores(ids)
	to_write := scoresToCSV(scores)
	w.WriteAll(to_write)
}

func main() {
	start := time.Now()

	lineLooperz("")
	// writeScores("basketball")

	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Println(elapsed)
}

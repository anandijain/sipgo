// credit montanaflynn/pget.go https://gist.github.com/montanaflynn/ea4b92ed640f790c4b9cee36046a5383
package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"
)

var lineHeaders = []string{"sport", "game_id", "a_team", "h_team", "num_markets", "a_ml", "h_ml", "draw_ml", "last_mod"}
var scoreHeaders = []string{"game_id", "a_team", "h_team", "period", "secs", "is_ticking", "a_pts", "h_pts", "status", "last_mod_score"}

var allHeaders = []string{"sport", "league", "subleague", "game_id", "a_team", "h_team", "num_markets", "a_ml",
	"h_ml", "draw_ml", "game_start", "last_mod", "period", "secs", "is_ticking", "a_pts",
	"h_pts", "status"}

var schema = `CREATE TABLE rows(
	id int NOT NULL AUTO_INCREMENT, 
	sport varchar(4), 
	league varchar(128), 
	subleague varchar(128), 
	game_id int, 
	a_team varchar(64), 
	h_team varchar(64), 
	num_markets int, 
	a_ml decimal, 
	h_ml decimal, 
	draw_ml decimal,  
	game_start int, 
	last_mod int,
	period int,
	secs int, 
	is_ticking boolean,
	a_pts int,
	h_pts int,
	status varchar(32),
	PRIMARY KEY (id))`

var lineRoot = "https://www.bovada.lv/services/sports/event/v2/events/A/description/"
var scoreRoot = "https://www.bovada.lv/services/sports/results/api/v1/scores/"

type concurrentResRow struct {
	index int
	res   Row
	err   error
}

// Line for CSV headers len 17
type Row struct {
	Sport      string
	League     string
	Subleague  string
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

func rowToCSV(r Row) []string {
	ret := []string{
		r.Sport,
		r.League,
		r.Subleague,
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
		cat := parsePaths(ev.Paths)
		// for i, p := range ev.Paths{
		// 	if p.PathType == "LEAGUE":
		// }
		for _, e := range es {
			r, null_row := makeLineToRow(e)
			if null_row == false {
				r.League = cat[1]
				r.Subleague = cat[2]
				rs[r.GameID] = r
			}
		}
	}

	return rs
}

func parsePaths(ps []Path) [3]string {
	// in order: sport, league, subleague
	var ret [3]string
	if len(ps) == 3 {
		ret[0] = ps[2].Description
		ret[1] = strings.Replace(ps[1].Description, ",", "", -1)
		ret[2] = strings.Replace(ps[0].Description, ",", "", -1)
	} else if len(ps) == 2 {
		ret[0] = ps[1].Description
		ret[1] = strings.Replace(ps[0].Description, ",", "", -1)
		ret[2] = ""
	}
	return ret
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
func makedb() *sql.DB {
	db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/hello")
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()
	err = db.Ping()
	if err != nil {
		fmt.Println("ruh roh")
	}

	_, err = db.Exec("CREATE DATABASE testDB")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Successfully created database..")
	}
	_, err = db.Exec("USE testDB")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("DB selected successfully..")
	}
	return db
}

var insertRowsQuery = `INSERT INTO rows 
(sport, league, subleague, game_id, a_team, h_team, num_markets, a_ml,
	h_ml, draw_ml, game_start, last_mod, period, secs, is_ticking, a_pts,
	h_pts, status)
	VALUES(?)`

func insertRows(db *sql.DB, rs []Row) {
	for _, r := range rs {
		stmt, err := db.Prepare(insertRowsQuery)
		if err != nil {
			log.Fatal(err)
		}
		res, err := stmt.Exec(
			reflect.ValueOf(r))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res)
	}
}

func makeRowsTable(db *sql.DB) error {
	stmt, err := db.Prepare(schema)
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = stmt.Exec()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Table created successfully..")
	}
	return err
}

func main() {
	// looperz("soccer", "rows.csv")
	db := makedb()

	fmt.Println(db)
}

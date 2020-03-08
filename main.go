// credit montanaflynn/pget.go https://gist.github.com/montanaflynn/ea4b92ed640f790c4b9cee36046a5383
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func getRows(s string) (map[int]Row, error) {
	ret, err := req(lineRoot + s)
	rs := parseLinesToRows(ret)
	return rs, err
}

func parsePaths(ps []Path) (Row, error) {
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

func parseLinesToRows(b []byte) map[int]Row {
	data := toJSON(b)
	rs := make(map[int]Row)
	for _, ev := range data {
		es := ev.Events
		categories := parsePaths(ev.Paths)

		for _, e := range es {
			r, null_row := makeLineToRow(e)
			if null_row == false {
				r.Country = categories.Country
				r.Region = categories.Region
				r.League = categories.League
				r.Comp = categories.Comp
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
	rs, _ := getRows(s)
	rs = addScores(rs, 16)
	return rs
}

func looperz(s string, fn string) {
	_, w := initCSV(fn, allHeaders)
	url := lineRoot + s
	prev := grabRows(url)
	initWrite := rowsToCSVFormat(prev)
	w.WriteAll(initWrite)
	cur := grabRows(url)

	for {
		diff := compRows(prev, cur)
		rowsToWrite := rowsToCSVFormat(diff)
		fmt.Println(len(rowsToWrite), "# of changes", time.Now())

		prev = cur
		cur = grabRows(url)

		w.WriteAll(rowsToWrite)
	}
}

func main() {
	// looperz("", "wleagues.csv")
	db := initDB("rows")
	fmt.Println(db)

	rs, _ := getRows("")
	insertRows(db, rs)
	db.Close()
}

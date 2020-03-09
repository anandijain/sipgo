// credit montanaflynn/pget.go https://gist.github.com/montanaflynn/ea4b92ed640f790c4b9cee36046a5383
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/mysql"
	_ "github.com/go-sql-driver/mysql"
)

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

func getRowsNoScore(s string) (map[int]Row, error) {
	ret, err := req(lineRoot + s)
	rs := parseLinesToRows(ret)
	return rs, err
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

func grabRows(s string) map[int]Row {
	rs, _ := getRowsNoScore(s)
	rs = addScores(rs, 16)
	return rs
}

func looperz(s string, fn string) {
	_, w := initCSV(fn, allHeaders)
	prev := grabRows(s)
	initWrite := rowsToCSVFormat(prev)
	w.WriteAll(initWrite)
	cur := grabRows(s)

	for {
		diff := compRows(prev, cur)
		rowsToWrite := rowsToCSVFormat(diff)
		fmt.Println(len(rowsToWrite), "# of changes", time.Now())

		prev = cur
		cur = grabRows(s)

		w.WriteAll(rowsToWrite)
	}
}

func testInsertDB(name string) {
	db := initCloudDB(name)
	fmt.Println(db)

	rs := grabRows("")
	insertRows(db, rs)
	db.Close()
}

func main() {
	// looperz("", "data.csv")
	testInsertDB("rows")
}

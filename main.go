// credit montanaflynn/pget.go https://gist.github.com/montanaflynn/ea4b92ed640f790c4b9cee36046a5383
// https://medium.com/dev-bits/making-concurrent-http-requests-in-go-programming-language-823b51bb1dc2
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/mysql"
	_ "github.com/go-sql-driver/mysql"
)

func req(s string) ([]byte, error) {
	res, _ := http.Get(s)
	// if httperr != nil {
	// 	fmt.Println("1")
	// 	fmt.Println(res)
	// } else {
	// 	fmt.Println("req error")
	// 	fmt.Println(res)
	// 	log.Fatal(httperr)
	// }
	ret, err := ioutil.ReadAll(res.Body)
	// fmt.Println(ret)
	res.Body.Close()
	return ret, err
}

func parseLinesToRows(b []byte) map[int]Row {
	data := toJSON(b)
	rs := make(map[int]Row)
	for _, ev := range data {
		es := ev.Events
		categories := parsePaths(ev.Paths)

		for _, e := range es {
			r, nullRow := makeLineToRow(e)
			if nullRow == false {
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

func getScores() map[int]shortScore {
	ret, _ := req(scoreRoot)
	shortScores := make(map[int]shortScore)
	scores := scoresFromBytes(ret)
	fmt.Println(scores)
	for _, s := range scores {
		for _, sc := range s.Scores {
			toAdd, nullRow := makeScore(sc)
			if nullRow != true {
				shortScores[toAdd.GameID] = toAdd
			}
		}
	}
	return shortScores
}

func getRows(s string) map[int]Row {
	rs, _ := getRowsNoScore(s)
	rs = addScores(rs, concurrencyLim)
	return rs
}

func getLines(s string) (map[int]Line, error) {
	ret, err := req(lineRoot + s)
	rs := parseLines(ret)
	return rs, err
}

func looperz(fn string) {
	_, w := initCSV(fn, allHeaders)
	prev := getRows("")
	w.WriteAll(rowsToCSVFormat(prev))
	cur := getRows("")

	for {
		diff := compRows(prev, cur)
		fmt.Println(len(diff), "# of changes", time.Now())

		prev = cur
		cur = getRows("")

		w.WriteAll(rowsToCSVFormat(diff))
	}
}

func loopDB(name string) {

	db := initCloudDB(name)
	stmt, _ := db.Prepare(insertRowsQuery)

	prev := getRows("")
	cur := getRows("")

	for {
		diff := compRows(prev, cur)
		fmt.Println(len(diff), "# of changes", time.Now())
		
		insertRows(db, diff, stmt)
		
		prev = cur
		cur = getRows("")
	}
}

func testInsertDB(name string) {
	db := initCloudDB(name)
	stmt, _ := db.Prepare(insertRowsQuery)
	
	rs := getRows("")
	insertRows(db, rs, stmt)
	db.Close()
}

func lineLooperz(s string) {
	_, w := initCSV("lines3.csv", lineHeaders)
	
	prev, _ := getLines("")
	cur, _ := getLines("")
	
	for {
		diff :=  make(map[int]Line)
		for id, v := range cur {
			if !reflect.DeepEqual(v, prev[id]) {
				diff[id] = v
			}
		}
		w.WriteAll(linesToCSV(diff))
		fmt.Println(len(diff), "# of changes", time.Now())
		
		prev = cur
		cur, _ = getLines("")
		time.Sleep(3000)
	}
	
}
func main() {
	// looperz("data.csv")

	// // rs := getRows("")
	// var ids []int
	// for id := range rs {
	// 	ids = append(ids, id)
	// }
	lineLooperz("")
	// t0 := time.Now()
	// ss := getScores()
	// t1 := time.Now()
	// fmt.Println(ss)
	// fmt.Println("getLines", len(ss), "# lines and scores in", t1.Sub(t0))
	// testInsertDB("rows")
	// loopDB("", "rows")
}

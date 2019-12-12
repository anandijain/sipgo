package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Row for CSV headers len 8
type Row struct {
	Sport      string
	GameID     int
	aTeam      string
	hTeam      string
	NumMarkets int
	aML        float64
	hML        float64
	aPS        float64
	hPS        float64
	aHC        float64
	hHC        float64
	gameStart  int
	LastMod    int
}

func makeRow(e Event) Row {
	var r Row
	r.Sport = e.Sport
	gameID, _ := strconv.Atoi(e.ID)
	// if err != nil {
	// 	fmt.Println("gameID weirdness", err)
	// 	gameID := ""
	// }
	r.GameID = gameID

	if e.AwayTeamFirst {
		r.aTeam = e.Competitors[0].Name
		r.hTeam = e.Competitors[1].Name
	} else {
		r.aTeam = e.Competitors[1].Name
		r.hTeam = e.Competitors[0].Name
	}
	mkts := e.DisplayGroups[0].Markets
	mainMkts := getMainMarkets(mkts)

	aML, _ := strconv.ParseFloat(mainMkts["Moneyline"].Outcomes[0].Price.Decimal, 64)
	hML, _ := strconv.ParseFloat(mainMkts["Moneyline"].Outcomes[1].Price.Decimal, 64)

	aPS, _ := strconv.ParseFloat(mainMkts["Point Spread"].Outcomes[0].Price.Decimal, 64)
	hPS, _ := strconv.ParseFloat(mainMkts["Point Spread"].Outcomes[1].Price.Decimal, 64)

	aHC, _ := strconv.ParseFloat(mainMkts["Point Spread"].Outcomes[0].Price.Handicap, 64)
	hHC, _ := strconv.ParseFloat(mainMkts["Point Spread"].Outcomes[1].Price.Handicap, 64)

	r.aML = aML
	r.hML = hML

	r.aPS = aPS
	r.hPS = hPS

	r.aHC = aHC
	r.hHC = hHC

	r.LastMod = e.LastModified
	r.gameStart = e.StartTime
	r.NumMarkets = e.NumMarkets

	return r
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

func getRows(b []byte) []Row {
	data := toJSON(b)
	var rs []Row
	compEvents := data[0].Events
	numEvents := len(compEvents)

	for i := 0; i < numEvents; i++ {
		r := makeRow(compEvents[i])
		rs = append(rs, r)
	}
	return rs
}

func getSport(s string) []Row {
	res, err := http.Get("https://www.bovada.lv/services/sports/event/v2/events/A/description/football/nfl")
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

	rs := getRows(ret)
	// w := csv.NewWriter(os.Stdout)

}

func main() {

	headers := `{sport,game_id,a_team,h_team,a_ml,h_ml,last_mod,num_markets}`
	fmt.Println(headers)

	for _, row := range rs {
		fmt.Println(row)
	}
	// fmt.Println(rs)

	// for i, record := range rs {
	// 	s := reflect.ValueOf(&record).Elem()
	// 	for i := 0; i < s.NumField(); i++ {
	// 		f := s.Field(i)
	// 		fmt.Printf("%d: %s %s = %v\n", i, f.Type(), f.Interface())
	// 		if err := w.Write(string(f)); err != nil {
	// 			log.Fatalln("error writing record to csv:", err)
	// 		}
	// 	}
	// }
}

func toJSON(b []byte) []Competition {
	retString := string(b)
	dec := json.NewDecoder(strings.NewReader(retString))
	var c []Competition
	for {
		if err := dec.Decode(&c); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
	}
	return c
}

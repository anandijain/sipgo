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

var sportsMap = map[string]string{
	"nba":  "basketball/nba",
	"nfl":  "football/nfl",
	"nhl":  "hockey/nhl",
	"mlb":  "baseball/mlb",
	"FOOT": "football/",
	"BASK": "basketball/",
}

func makeRow(e Event) (Row, bool) {
	var r Row
	var null_row = false
	r.Sport = e.Sport
	gameID, _ := strconv.Atoi(e.ID)
	// if err != nil {
	// 	fmt.Println("gameID weirdness", err)
	// 	gameID := ""
	// }
	r.GameID = gameID

	// fmt.Println(e.Competitors)
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

	// var mkts []Market
	// for _, dg := range e.DisplayGroups {
	// 	for  _, value := range getMainMarkets(dg.Markets) {
	// 	   mkts = append(value, mkts)
	// 	}

	// }
	mkts := e.DisplayGroups[0].Markets
	mainMkts := getMainMarkets(mkts)

	mls := mainMkts["Moneyline"].Outcomes
	if len(mls) == 2 {
		aML, err := strconv.ParseFloat(mls[0].Price.Decimal, 64)
		if err != nil {
			fmt.Println("closed", err)
			r.aML = 0
			null_row = true
		}
		hML, err := strconv.ParseFloat(mls[1].Price.Decimal, 64)
		if err != nil {
			fmt.Println("closed", err)
			r.hML = 0
			null_row = true

		}

		r.aML = aML
		r.hML = hML

	} else {
		r.aML = 0
		r.hML = 0
		null_row = true
	}
	if null_row {
		return r, true
	}
	aPS, _ := strconv.ParseFloat(mainMkts["Point Spread"].Outcomes[0].Price.Decimal, 64)
	hPS, _ := strconv.ParseFloat(mainMkts["Point Spread"].Outcomes[1].Price.Decimal, 64)

	aHC, _ := strconv.ParseFloat(mainMkts["Point Spread"].Outcomes[0].Price.Handicap, 64)
	hHC, _ := strconv.ParseFloat(mainMkts["Point Spread"].Outcomes[1].Price.Handicap, 64)

	r.aPS = aPS
	r.hPS = hPS

	r.aHC = aHC
	r.hHC = hHC

	r.LastMod = e.LastModified
	r.gameStart = e.StartTime
	r.NumMarkets = e.NumMarkets
	if r.aTeam == "" {
		fmt.Println("null row")
		null_row = true
	}
	return r, null_row
}

func makeScore(s Score) shortScore {
	var r shortScore
	r.Period = s.Clock.PeriodNumber
	r.Seconds = s.Clock.RelativeGameTimeInSecs
	r.IsTicking = s.Clock.IsTicking
	// r.NumberOfPeriods = s.Clock.NumberOfPeriods
	if s.Competitors[0].Name == "" {
		fmt.Println("broke")
	}
	r.aTeam = s.Competitors[0].Name
	r.hTeam = s.Competitors[1].Name
	r.aPts = s.LatestScore.Visitor
	r.hPts = s.LatestScore.Home
	r.Status = s.GameStatus
	r.lastMod = s.LastUpdated
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

func scoreToJSON(b []byte) Score {
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

func getSport(s string) []Row {
	res, err := http.Get("https://www.bovada.lv/services/sports/event/v2/events/A/description/" + sportsMap[s])
	fmt.Println(res)
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
	return rs
}

func getRows(b []byte) []Row {
	data := toJSON(b)
	var rs []Row
	// var events []Event

	for _, ev := range data {
		es := ev.Events
		for _, e := range es {
			// events = append(e, events)
			r, null_row := makeRow(e)
			fmt.Println(r)
			// s := getScore(e.ID)
			// fmt.Println(s)
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


func main() {
	lines, err := os.Create("lines.csv")
	scores, err := os.Create("scores.csv")
	
	linesWriter := csv.NewWriter(lines)
	scoresWriter := csv.NewWriter(scores)
	
	defer linesWriter.Flush()
	defer scoresWriter.Flush()

	headers := `{sport,game_id,a_team,h_team,a_ml,h_ml,last_mod,num_markets}`
	scoreHeaders := `{game_id,a_team,h_team,period,secs,is_ticking,a_pts,h_pts,status,last_mod}`

	fmt.Println(headers)
	fmt.Println(scoreHeaders)

	nba := getSport("nba")
	ids := idsFromRows(nba)
	scores = getScores(ids)

    for _, value := range  {
        err := writer.Write(value)
        checkError("Cannot write to file", err)
    }
	checkError("Cannot create file", err)
    defer file.Close()

}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func main() {
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

	data := toJSON(ret)
	compEvents := data[0].Events
	for i := 0; i < len(compEvents); i++ {
		game := compEvents[i]
		fmt.Println(game.ID, game.Description)
		dgs := game.DisplayGroups
		for j := 0; j < len(dgs); j++ {
			dg := dgs[j]
			mkts := dg.Markets
			for k := 0; k < len(mkts); k++ {
				mkt := mkts[k]
				fmt.Println(mkt.Description)
				outcomes := mkt.Outcomes
				for l := 0; l < len(outcomes); l++ {
					outcome := outcomes[l]
					fmt.Println(mkt.Description, outcome.Price.American)
				}
			}
		}
	}

	// json, err := json.MarshalIndent(data, "  ", "    ")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(string(json))
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

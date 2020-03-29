package main

import (
	"fmt"
	"testing"
	"time"
)

func evalConcurrentSpeedup() {
	start := time.Now()
	data := parallelGetRows(supportedSports, concurrencyLim)
	t := time.Now()
	fmt.Println("concurrent got", len(data), "in", t.Sub(start))

	start2 := time.Now()
	prev := getRows("")
	cur := getRows("")
	diff := compRows(prev, cur)

	t2 := time.Now()
	fmt.Println("not concurrent got", len(diff), "in", t2.Sub(start2))

}

// func TestSeparateConcurrent(t *testing.T) {
// 	t0 := time.Now()
// 	rs := separateConcurrent(t.sports, t.concurrencyLimit)
// 	t1 := time.Now()
// 	fmt.Println(len(rs), "# lines and scores in", t1.Sub(t0))
// }

// func TestSeparateConcurrent(t *testing.T) {

// }

func TestTableSeparateConcurrent(t *testing.T) {
	var tests = []struct {
		sports []string
		lim    int
	}{
		{[]string{"football", "soccer", "basketball"}, 10},
		{[]string{"football", "soccer", "basketball"}, 5},
		{[]string{"esports", "snooker", "cricket"}, 10},
		{[]string{"esports", "snooker", "cricket"}, 5},
	}
	for _, test := range tests {
		t0 := time.Now()
		rs := separateConcurrent(test.sports, test.lim)
		t1 := time.Now()
		fmt.Println(len(rs), "# lines and scores in", t1.Sub(t0))
	}
}

func parallelDiffs(sports []string) map[int]Row {
	prev := parallelGetRows(sports, concurrencyLim)
	cur := parallelGetRows(sports, concurrencyLim)
	diffs := getDiffs(prev, cur)
	return diffs
}

// t0 := time.Now()
// rs := separateConcurrent(supportedSports, 10)
// t1 := time.Now()
// fmt.Println("sepCon", len(rs), "# lines and scores in", t1.Sub(t0))
// t0 = time.Now()
// prs := parallelGetRows(supportedSports, 10)
// t1 = time.Now()
// fmt.Println("parGet in", t1.Sub(t0))
// for _, psport := range prs {
// 	fmt.Println("parGet", len(psport.res), "# lines and scores in", t1.Sub(t0))
// }
// t0 = time.Now()
// grs := make(map[int]Row)
// for _, s := range supportedSports {
// 	for id, r := range getRows(s) {
// 		grs[id] = r
// 	}
// }
// t1 = time.Now()
// fmt.Println("getRows", len(grs), "# lines and scores in", t1.Sub(t0))

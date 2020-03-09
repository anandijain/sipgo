package main

import (
	"database/sql"
	"fmt"
	"log"
)

func initDB(name string) *sql.DB {
	db, err := sql.Open("mysql", "root:@/rows")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
	}
	useDB(db, name) // database: rows
	useDB(db, name) // table rows

	return db
}

func createDB(db *sql.DB, name string) {
	_, err := db.Exec("CREATE DATABASE testDB")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Successfully created database..")
	}
}

func useDB(db *sql.DB, name string) {
	_, err := db.Exec("USE " + name)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("DB selected successfully..")
	}
}

var insertRowsQuery = `INSERT INTO rows 
	(
		game_id, sport, league, comp, country, region, a_team, 
		h_team, num_markets, a_ml, h_ml, draw_ml, game_start, 
		last_mod, period, secs, is_ticking, a_pts, h_pts, status
	) 
	VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

var insertRowsQuery2 = `INSERT INTO rows 
(game_id, sport, a_team, h_team, num_markets, a_ml, h_ml, draw_ml)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)`

func insertRows(db *sql.DB, rs map[int]Row) {
	for _, r := range rs {
		stmt, err := db.Prepare(insertRowsQuery)
		if err != nil {
			fmt.Println("prep error")
			log.Fatal(err)
		}

		// fmt.Println(fmt.Sprint(r.aTeam))
		// fmt.Println(r.aML)
		// if r.aPts == '' {
		// 	r.aPts = 0
		// }
		// if r.hPts == '' {
		// 	r.hPts = 0
		// }
		// fmt.Println(r.gameStart)
		_, err = stmt.Exec(
			r.GameID,
			r.Sport,
			r.League,
			r.Comp,
			r.Country,
			r.Region,
			fmt.Sprint(r.aTeam),
			fmt.Sprint(r.hTeam),
			r.NumMarkets,
			r.aML,
			r.hML,
			r.drawML,
			r.gameStart,
			r.LastMod,
			r.Period,
			r.Seconds,
			r.IsTicking,
			r.aPts,
			r.hPts,
			r.Status)
		if err != nil {
			fmt.Println("write error", err)
			// log.Fatal("write error:", err)
		}
		// fmt.Println(res)
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

func testInsertDB() {
	db := initDB("rows")
	fmt.Println(db)

	rs := grabRows("")
	insertRows(db, rs)
	db.Close()
}

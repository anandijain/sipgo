# writes odds for bov to CSV

```go
go run types.go utils.go main.go
```

Collects the moneyline of ~500 events and the score of live games.

## benchmarks 

at ~20 mbps 

* requesting 1564 (431 with NA drop) lines and scores takes ~~~15 mins~~ ~25 seconds
* 1670 (1669 with NA drop) of lines took ~20 seconds


## data breakdown

odds:

    sport: FOOT, shape: (4, 8)
    sport: CRIC, shape: (10, 8)
    sport: BASE, shape: (9, 8)
    sport: BASK, shape: (70, 8)
    sport: BOXI, shape: (35, 8)
    sport: TENN, shape: (141, 8)
    sport: SOCC, shape: (921, 8)
    sport: HAND, shape: (37, 8)
    sport: VOLL, shape: (9, 8)
    sport: MMA, shape: (63, 8)
    sport: FUTS, shape: (1, 8)
    sport: TABL, shape: (24, 8)
    sport: HCKY, shape: (61, 8)
    sport: AURL, shape: (7, 8)
    sport: ESPT, shape: (211, 8)
    sport: RUGU, shape: (24, 8)
    sport: RUGL, shape: (14, 8)
    sport: SNOO, shape: (8, 8)
    sport: DART, shape: (21, 8)

## TODO

* there is a bottleneck in requesting the entire /sports/ page at once, to fix this concurrently request each sport individually, separate live/pregame for req()
* misc sports fix (tennis scores, boxing/mma/cricket odds)

## completed

* cloud sql upload
* added region country and competition
* added league and round info to Row
* only writes changes
* 3-way ML fix
* scores using semaphores/concurrency

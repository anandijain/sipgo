# writes odds for bov to CSV

```go
go run types.go utils.go main.go
```

Collects the moneyline of ~500 events and the score of live games.

## TODO

* only grab score of live games, sorting live games by num_markets
* cloud sql upload

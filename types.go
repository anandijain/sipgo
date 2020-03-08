package main

type concurrentResRow struct {
	index int
	res   Row
	err   error
}

// Line for CSV headers len 17
type Row struct {
	GameID     int
	Sport      string
	League     string
	Comp       string
	Country    string
	Region     string
	aTeam      string
	hTeam      string
	NumMarkets int
	aML        float64
	hML        float64
	drawML     float64
	gameStart  int
	LastMod    int
	Period     int
	Seconds    int
	IsTicking  bool
	aPts       int
	hPts       int
	Status     string
	lastMod    string
}

// Sport : specifying game
type Sport struct {
	Competitions []Competition `json:""`
}

// Competition : specifying game
type Competition struct {
	Events []Event `json:"events"`
	Paths  []Path  `json:"path"`
}

// Event : specifying game
type Event struct {
	ID            string         `json:"id"`
	Description   string         `json:"description"`
	EventType     string         `json:"type"`
	Link          string         `json:"link"`
	Status        string         `json:"status"`
	Sport         string         `json:"sport"`
	StartTime     int            `json:"startTime"`
	Live          bool           `json:"live"`
	AwayTeamFirst bool           `json:"awayTeamFirst"`
	DenySameGame  string         `json:"denySameGame"`
	TeaserAllowed bool           `json:"teaserAllowed"`
	CompetitionID string         `json:"competitionId"`
	Notes         string         `json:"notes"`
	NumMarkets    int            `json:"numMarkets"`
	LastModified  int            `json:"lastModified"`
	Competitors   []Competitor   `json:"competitors"`
	DisplayGroups []DisplayGroup `json:"displayGroups"`
}

// DisplayGroup : specifying game
type DisplayGroup struct {
	ID            string   `json:"id"`
	Description   string   `json:"description"`
	DefaultType   bool     `json:"defaultType"`
	AlternateType bool     `json:"alternateType"`
	Markets       []Market `json:"markets"`
	Order         int      `json:"order"`
}

// Competitor : specifying game
type Competitor struct {
	Home bool   `json:"home"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Market : specifying game
type Market struct {
	ID           string    `json:"id"`
	Description  string    `json:"description"`
	Key          string    `json:"key"`
	MarketTypeID string    `json:"marketTypeId"`
	Status       string    `json:"status"`
	SingleOnly   bool      `json:"singleOnly"`
	Notes        string    `json:"notes"`
	Period       Period    `json:"period"`
	Outcomes     []Outcome `json:"outcomes"`
}

// Outcome : specifying game
type Outcome struct {
	ID           string `json:"id"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	OutcomeType  string `json:"outcomeType"`
	CompetitorID string `json:"competitorId"`
	Price        Price  `json:"price"`
}

// Price : specifying game
type Price struct {
	ID         string `json:"id"`
	Handicap   string `json:"handicap"`
	American   string `json:"american"`
	Decimal    string `json:"decimal"`
	Fractional string `json:"fractional"`
	Malay      string `json:"malay"`
	Indonesian string `json:"indonesian"`
	Hongkong   string `json:"hongkong"`
}

// Period : specifying game
type Period struct {
	ID           string `json:"id"`
	Description  string `json:"description"`
	Abbreviation string `json:"abbreviation"`
	Live         bool   `json:"live"`
	Main         bool   `json:"main"`
}

// Path : specifying path
type Path struct {
	ID          string `json:"id"`
	Link        string `json:"link"`
	Description string `json:"description"`
	PathType    string `json:"type"`
	SportCode   string `json:"sportCode"`
	Order       int    `json:"order"`
	Leaf        bool   `json:"leaf"`
	Current     bool   `json:"current"`
}

// ResultStruct : give json to this
type ResultStruct struct {
	Result []map[string]Competition
}

// ResultList : give json to this
type ResultList struct {
	ResultL []Competition
}

type Clock struct {
	Period                 string `json:"period"`
	PeriodNumber           int    `json:"periodNumber"`
	GameTime               string `json:"gameTime"`
	RelativeGameTimeInSecs int    `json:"relativeGameTimeInSecs"`
	Direction              string `json:"direction"`
	NumberOfPeriods        int    `json:"numberOfPeriods"`
	IsTicking              bool   `json:"isTicking"`
}

type scoreCompetitor struct {
	Name          string      `json:"name"`
	Nickname      string      `json:"nickname"`
	Abbreviation  string      `json:"abbreviation"`
	Type          string      `json:"type"`
	HomeOrVisitor string      `json:"homeOrVisitor"`
	TeamID        interface{} `json:"teamId"`
}

type latestScore struct {
	Home    string `json:"home"`
	Visitor string `json:"visitor"`
}
type altIds struct {
	BGS int `json:"BGS"`
}

type Score struct {
	EventID                 int               `json:"eventId"`
	EventSource             string            `json:"eventSource"`
	AltIds                  altIds            `json:"altIds"`
	ScoreboardAvailable     bool              `json:"scoreboardAvailable"`
	Sport                   string            `json:"sport"`
	LatestScore             latestScore       `json:"latestScore"`
	Clock                   Clock             `json:"clock"`
	Competitors             []scoreCompetitor `json:"competitors"`
	GameStatus              string            `json:"gameStatus"`
	KeyEvents               []interface{}     `json:"keyEvents"`
	LastUpdated             string            `json:"lastUpdated"`
	EventDescription        string            `json:"eventDescription"`
	DisplayVisitorTeamFirst bool              `json:"displayVisitorTeamFirst"`
}

type concurrentRes struct {
	index int
	res   shortScore
	err   error
}

// Line for CSV headers len 10
type Line struct {
	Sport      string
	GameID     int
	aTeam      string
	hTeam      string
	NumMarkets int
	aML        float64
	hML        float64
	drawML     float64
	gameStart  int
	LastMod    int
}

// score for CSV len 10
type shortScore struct {
	GameID    int
	aTeam     string
	hTeam     string
	Period    int
	Seconds   int
	IsTicking bool
	aPts      string
	hPts      string
	Status    string
	lastMod   string
}

// https://shikenso.com/api/v2/tournaments?token=1111

type ShikensoTeam struct {
	Name          string `json:"name"`
	Abbrev        string `json:"abbrev"`
	TeamLogoLarge string `json:"team_logo_large"`
	TeamLogo      string `json:"team_logo"`
}

type Tourney []struct {
	WebsiteLink string `json:"website_link,omitempty"`
	Country     string `json:"country"`
	Image       string `json:"image"`
	Teams       []struct {
		Players []struct {
		} `json:"players"`
	} `json:"teams"`
	Prizepoolsum string `json:"prizepoolsum"`
	Start        int    `json:"start"`
	WikiLink     string `json:"wiki_link"`
	GameTitle    string `json:"game_title"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	TournamentID int    `json:"tournament_id"`
	Tier         string `json:"tier"`
	Series       []struct {
		Name string `json:"name"`
	} `json:"series,omitempty"`
	Prizepool []PrizePool
	End       int       `json:"end"`
	Sponsors  []Sponsor `json:"sponsors,omitempty"`
}

type PrizePool struct {
	Place int          `json:"place"`
	Team  ShikensoTeam `json:"team"`
	Prize int          `json:"prize"`
}

type Sponsor struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

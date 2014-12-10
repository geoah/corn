package main

type EzTvSeries struct {
	V        float64 `json:"__v"`
	ID       string  `json:"_id"`
	AirDay   string  `json:"air_day"`
	AirTime  string  `json:"air_time"`
	Country  string  `json:"country"`
	Episodes []struct {
		DateBased  bool   `json:"date_based"`
		Episode    uint64 `json:"episode"`
		FirstAired uint64 `json:"first_aired"`
		Overview   string `json:"overview"`
		Season     uint64 `json:"season"`
		Title      string `json:"title"`
		Torrents   struct {
			Sd struct {
				Peers uint64 `json:"peers"`
				Seeds uint64 `json:"seeds"`
				URL   string `json:"url"`
			} `json:"0"`
			Sd480p struct {
				Peers uint64 `json:"peers"`
				Seeds uint64 `json:"seeds"`
				URL   string `json:"url"`
			} `json:"480p"`
			Hd720p struct {
				Peers uint64 `json:"peers"`
				Seeds uint64 `json:"seeds"`
				URL   string `json:"url"`
			} `json:"720p"`
		} `json:"torrents"`
		TvdbID  uint64 `json:"tvdb_id"`
		Watched struct {
			Watched bool `json:"watched"`
		} `json:"watched"`
	} `json:"episodes"`
	Genres []string `json:"genres"`
	Images struct {
		Banner string `json:"banner"`
		Fanart string `json:"fanart"`
		Poster string `json:"poster"`
	} `json:"images"`
	ImdbID      string `json:"imdb_id"`
	LastUpdated uint64 `json:"last_updated"`
	Network     string `json:"network"`
	NumSeasons  uint64 `json:"num_seasons"`
	Rating      struct {
		Hated      float64 `json:"hated"`
		Loved      float64 `json:"loved"`
		Percentage float64 `json:"percentage"`
		Votes      float64 `json:"votes"`
	} `json:"rating"`
	Runtime  string `json:"runtime"`
	Slug     string `json:"slug"`
	Status   string `json:"status"`
	Synopsis string `json:"synopsis"`
	Title    string `json:"title"`
	TvdbID   string `json:"tvdb_id"`
	Year     string `json:"year"`
}

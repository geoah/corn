package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/martini-contrib/encoder"
	"github.com/garfunkel/go-tvdb"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Episode struct {
	ID             uint64
	EpisodeName    string
	EpisodeNumber  uint64
	FirstAired     string
	ImdbID         string
	Language       string
	SeasonNumber   uint64
	LastUpdated    string
	SeasonID       uint64
	SeriesID       uint64
	HasAired       bool
	LocalFilename  string
	LocalExists    bool
	LocalQuality   string
	TorrentQuality string
	TorrentLink    string
}

// type SeasonEpisode struct {
// 	Season  uint64
// 	Episode uint64
// }

type Series struct {
	ID          uint64
	Matched     bool
	ImdbID      string
	Status      string
	SeriesID    string
	SeriesName  string
	Language    string
	LastUpdated string
	Episodes    map[string]*Episode
	LocalName   string
	LocalPath   string
	Poster      string
}

// Get basic information from tvdbcom
func (s *Series) fetchInfo() {
	// TODO Get just directory instead of full path if LocalPath is absolute
	seriesListTvDb, err := tvdb.GetSeries(s.LocalName)
	if err != nil || len(seriesListTvDb.Series) == 0 {
		s.Matched = false
		// fmt.Println("Could not match")
	} else {
		// series := *seriesListTvDb.Series[0]
		series, err := tvdb.GetSeriesByID(seriesListTvDb.Series[0].ID)
		if err != nil {
			s.Matched = false
		} else {
			s.Matched = true
			s.ID = series.ID
			s.ImdbID = series.ImdbID
			s.Status = series.Status
			s.SeriesName = series.SeriesName
			s.Language = series.Language
			s.LastUpdated = series.LastUpdated
			s.Poster = series.Poster
		}
	}
}

// Get detailed information from tvdbcom
func (s *Series) fetchDetails() {
	if s.Matched == true {
		seriesListTvDb, err := tvdb.GetSeries(s.LocalName)
		if err != nil || len(seriesListTvDb.Series) == 0 {
			// fmt.Println("Could not match")
			return
		}
		series := *seriesListTvDb.Series[0]
		series.GetDetail()
		// s.Episodes = make(map[SeasonEpisode]*Episode)
		s.Episodes = make(map[string]*Episode)

		for _, seasonEpisodes := range series.Seasons {
			for _, episode := range seasonEpisodes {
				episodeSimple := Episode{}

				episodeSimple.ID = episode.ID
				episodeSimple.EpisodeName = episode.EpisodeName
				episodeSimple.EpisodeNumber = episode.EpisodeNumber
				episodeSimple.FirstAired = episode.FirstAired
				episodeSimple.ImdbID = episode.ImdbID
				episodeSimple.Language = episode.Language
				episodeSimple.SeasonNumber = episode.SeasonNumber
				episodeSimple.LastUpdated = episode.LastUpdated
				episodeSimple.SeasonID = episode.SeasonID
				episodeSimple.SeriesID = episode.SeriesID

				if episode.FirstAired == "" {
					// fmt.Println("Missing first aired.")
				} else {
					aired, err := time.Parse("2006-01-02", episode.FirstAired)
					if err != nil {
						fmt.Println("Could not parse first aired.", err)
					} else {
						if aired.Before(time.Now()) {
							episodeSimple.HasAired = true
							// fmt.Println(series.SeriesName, "Season", episode.SeasonNumber, "Episode", episode.EpisodeNumber, "aired", episode.FirstAired)
						} else {
							episodeSimple.HasAired = false
							// fmt.Println(series.SeriesName, "Season", episode.SeasonNumber, "Episode", episode.EpisodeNumber, "not yet aired, airing on", episode.FirstAired)
						}
					}
				}
				// s.Episodes[SeasonEpisode{episode.SeasonNumber, episode.EpisodeNumber}] = &episodeSimple
				s.Episodes[fmt.Sprintf("%d_%d", episode.SeasonNumber, episode.EpisodeNumber)] = &episodeSimple
			}
		}
	} else {
		// TODO Log Error
	}
}

func (s *Series) CheckForExistingEpisodes() {
	patterns := []*regexp.Regexp{
		regexp.MustCompile("[Ss]([0-9]+)[][ ._-]*[Ee]([0-9]+)([^\\/]*)$"),
		regexp.MustCompile(`[\\/\._ \[\(-]([0-9]+)x([0-9]+)([^\\/]*)$`),
	}
	filepath.Walk(s.LocalPath, func(filePath string, f os.FileInfo, err error) error {
		if err != nil {
			// TODO Log Error
			return err
		}
		if !f.IsDir() && !strings.HasPrefix(f.Name(), ".") {
			for _, pattern := range patterns {
				matches := pattern.FindAllStringSubmatch(f.Name(), -1)
				if len(matches) > 0 && len(matches[0]) > 0 {
					season, _ := strconv.ParseUint(matches[0][1], 10, 64)
					episode, _ := strconv.ParseUint(matches[0][2], 10, 64)
					seasonEpisode := fmt.Sprintf("%d_%d", season, episode)
					if _, ok := s.Episodes[seasonEpisode]; ok {
						s.Episodes[seasonEpisode].LocalExists = true
						s.Episodes[seasonEpisode].LocalFilename = filepath.Join(s.LocalPath, f.Name())
					}
					break
				}
			}
		}
		return nil
	})
}

func (s *Series) FetchTorrentLinks() {
	url := "http://eztvapi.re/show/" + s.ImdbID
	res, err := http.Get(url)
	// TODO Check for errors
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	// TODO Check for errors
	var ezTvSeries EzTvSeries
	err = json.Unmarshal(body, &ezTvSeries)
	if err != nil {
		// TODO Log Error
		// fmt.Println("Somethign went wrong.")
		// fmt.Printf("%T\n%s\n%#v\n", err, err, err)
		// switch v := err.(type) {
		// case *json.SyntaxError:
		// fmt.Println(string(body[v.Offset-40 : v.Offset]))
		// }
	} else {
		for _, episode := range ezTvSeries.Episodes {
			var torrentLink string
			var torrentQuality string
			if episode.Torrents.Hd720p.URL != "" {
				torrentQuality = "720p"
				torrentLink = episode.Torrents.Hd720p.URL
			} else if episode.Torrents.Sd480p.URL != "" {
				torrentQuality = "480p"
				torrentLink = episode.Torrents.Sd480p.URL
			} else {
				torrentQuality = "sdtv"
				torrentLink = episode.Torrents.Sd.URL
			}
			// k :=SeasonEpisode{episode.Season, episode.Episode}
			k := fmt.Sprintf("%d_%d", episode.Season, episode.Episode)
			if torrentLink != "" && s.Episodes[k] != nil {
				s.Episodes[k].TorrentQuality = torrentQuality
				s.Episodes[k].TorrentLink = torrentLink
			}
		}
	}
}

func (s *Series) PrintResults() {
	for _, episode := range s.Episodes {
		fmt.Printf("[%s][%s] Season %d Episode %d ", s.LocalName, s.SeriesName, episode.SeasonNumber, episode.EpisodeNumber)
		if episode.HasAired {
			if episode.LocalExists {
				fmt.Printf("is available locally.\n")
			} else {
				fmt.Printf("is unavailable locally")
				if episode.TorrentLink != "" {
					fmt.Printf(" but I got a magnet link for %s", episode.TorrentQuality)
				}
				fmt.Printf(".\n")
			}
		} else {
			fmt.Printf("has not aired yet.\n")
		}
	}
}

func (s *Series) PrintJsonResults() {
	b, err := json.Marshal(s)
	if err != nil {
		fmt.Println("error:", err)
	}
	os.Stdout.Write(b)
}

// Martini instance
var m *martini.Martini
var store Store

// Create config struct to hold random things
var config struct {
	tvPath string
}

func init() {
	// Initialize store
	store = &SeriesStore{
		m: make(map[uint64]*Series),
	}

	// Initialize martini
	m = martini.New()

	// Setup martini middleware
	m.Use(martini.Recovery())
	m.Use(martini.Logger())

	// Setup routes
	r := martini.NewRouter()
	r.Get(`/series`, GetAllSeries)
	r.Get(`/series/:id`, GetSeries)

	// Allow CORS
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	// Other stuff
	m.Use(func(c martini.Context, w http.ResponseWriter) {
		// Inject JSON Encoder
		c.MapTo(encoder.JsonEncoder{}, (*encoder.Encoder)(nil))
		// Force Content-Type
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	})
	// Inject database
	m.MapTo(store, (*Store)(nil))
	// Add the router action
	m.Action(r.Handle)
}

func main() {
	// Check tvpath argument
	if len(os.Args) == 1 {
		fmt.Println("Missing tv directory path.")
		return
	}

	// Get tvpath from args
	config.tvPath = filepath.Clean(os.Args[1])

	// Populate series from tvpath
	PopSeries()

	// Startup HTTP server
	if err := http.ListenAndServe(":8000", m); err != nil {
		log.Fatal(err)
	}
}

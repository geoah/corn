package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/garfunkel/go-tvdb"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
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

type SeasonEpisode struct {
	Season  uint64
	Episode uint64
}

type Series struct {
	ID          uint64
	Matched     bool
	ImdbID      string
	Status      string
	SeriesID    string
	SeriesName  string
	Language    string
	LastUpdated string
	Episodes    map[SeasonEpisode]*Episode
	LocalName   string
	LocalPath   string
}

// getSeriesInfo from tvdbcom
func (s *Series) fetchInfo() {
	// TODO Get just directory instead of full path if LocalPath is absolute
	seriesListTvDb, err := tvdb.GetSeries(s.LocalName)
	if err != nil {
		fmt.Println("Could not match")
		return
	}

	if len(seriesListTvDb.Series) > 0 {
		s.Matched = true
		series := *seriesListTvDb.Series[0]
		series.GetDetail()

		s.ID = series.ID
		s.ImdbID = series.ImdbID
		s.Status = series.Status
		s.SeriesName = series.SeriesName
		s.Language = series.Language
		s.LastUpdated = series.LastUpdated
		s.Episodes = make(map[SeasonEpisode]*Episode)

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
				s.Episodes[SeasonEpisode{episode.SeasonNumber, episode.EpisodeNumber}] = &episodeSimple
			}
		}
	} else {
		// TODO Log Error
	}
}

func (s *Series) CheckForExistingEpisodes() {
	regOne := regexp.MustCompile("[Ss]([0-9]+)[][ ._-]*[Ee]([0-9]+)([^\\/]*)$")

	filepath.Walk(s.LocalPath, func(filePath string, f os.FileInfo, err error) error {
		if err != nil {
			// TODO Log Error
			return err
		}
		if !f.IsDir() && !strings.HasPrefix(f.Name(), ".") {
			res := regOne.FindAllStringSubmatch(f.Name(), -1)
			if len(res) > 0 && len(res[0]) > 0 {
				season, _ := strconv.ParseUint(res[0][1], 10, 64)
				episode, _ := strconv.ParseUint(res[0][2], 10, 64)
				if _, ok := s.Episodes[SeasonEpisode{season, episode}]; ok {
					s.Episodes[SeasonEpisode{season, episode}].LocalExists = true
					s.Episodes[SeasonEpisode{season, episode}].LocalFilename = filepath.Join(s.LocalPath, f.Name())
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
			if torrentLink != "" && s.Episodes[SeasonEpisode{episode.Season, episode.Episode}] != nil {
				s.Episodes[SeasonEpisode{episode.Season, episode.Episode}].TorrentQuality = torrentQuality
				s.Episodes[SeasonEpisode{episode.Season, episode.Episode}].TorrentLink = torrentLink
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

func main() {
	app := cli.NewApp()

	app.Action = func(c *cli.Context) {
		if len(c.Args()) == 0 {
			fmt.Println("Missing tv directory path.")
			return
		}

		// Get tvpath from args
		tvpath := filepath.Clean(c.Args()[0])

		// Open tvpath directory
		dir, err := os.Open(tvpath)
		if err != nil {
			return
		}
		defer dir.Close()

		var wg sync.WaitGroup
		// Loop tvpath for folders and try to match them with series from TvDB
		files, err := ioutil.ReadDir(tvpath)
		if err != nil {
			fmt.Println("Could not list directory contents", err)
		} else {
			for _, folder := range files {
				if folder.IsDir() && !strings.HasPrefix(folder.Name(), ".") {
					// Add to queue
					wg.Add(1)
					go func(tvpath string, folderName string) {
						var series Series = Series{}
						series.LocalName = folderName
						series.LocalPath = filepath.Join(tvpath, folderName)
						series.fetchInfo()
						if series.Matched == true {
							series.CheckForExistingEpisodes()
							series.FetchTorrentLinks()
							series.PrintResults()
						}
						// Remove from queue
						wg.Done()
					}(tvpath, folder.Name())
				}
			}
		}
		// Wait for queue to be completed
		wg.Wait()
	}

	app.Run(os.Args)
}

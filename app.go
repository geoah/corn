package main

import (
	// "encoding/json"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/garfunkel/go-tvdb"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Episode struct {
	ID            uint64
	EpisodeName   string
	EpisodeNumber uint64
	FirstAired    string
	ImdbID        string
	Language      string
	SeasonNumber  uint64
	LastUpdated   string
	SeasonID      uint64
	SeriesID      uint64
	HasAired      bool
	LocalFilename string
	LocalExists   bool
}

type SeasonEpisode struct {
	Season  uint64
	Episode uint64
}

type Series struct {
	ID          uint64
	ImdbID      string
	Status      string
	SeriesID    string
	SeriesName  string
	Language    string
	LastUpdated string
	Episodes    map[SeasonEpisode]*Episode
	LocalPath   string
}

// getSeriesInfo from tvdbcom
func getSeriesInfo(seriesName string) (Series, error) {
	seriesListTvDb, err := tvdb.GetSeries(seriesName)

	if err != nil {
		return Series{}, err
	}

	if len(seriesListTvDb.Series) > 0 {
		var series = *seriesListTvDb.Series[0]
		series.GetDetail()

		seriesSimple := Series{}
		seriesSimple.ID = series.ID
		seriesSimple.ImdbID = series.ImdbID
		seriesSimple.Status = series.Status
		seriesSimple.SeriesName = series.SeriesName
		seriesSimple.Language = series.Language
		seriesSimple.LastUpdated = series.LastUpdated
		seriesSimple.Episodes = make(map[SeasonEpisode]*Episode)

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
				// seriesSimple.Seasons[episode.SeasonNumber] = append(seriesSimple.Seasons[episode.SeasonNumber], &episodeSimple)
				seriesSimple.Episodes[SeasonEpisode{episode.SeasonNumber, episode.EpisodeNumber}] = &episodeSimple
			}
		}
		return seriesSimple, nil
	} else {
		return Series{}, errors.New("Not found")
	}
}

func (s *Series) CheckForExistingEpisodes() {
	regOne := regexp.MustCompile("[Ss]([0-9]+)[][ ._-]*[Ee]([0-9]+)([^\\/]*)$")

	// err :=
	filepath.Walk(s.LocalPath, func(path string, file os.FileInfo, err error) error {
		if !file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
			res := regOne.FindAllStringSubmatch(file.Name(), -1)
			season, _ := strconv.ParseUint(res[0][1], 10, 64)
			episode, _ := strconv.ParseUint(res[0][2], 10, 64)
			if _, ok := s.Episodes[SeasonEpisode{season, episode}]; ok {
				s.Episodes[SeasonEpisode{season, episode}].LocalExists = true
				s.Episodes[SeasonEpisode{season, episode}].LocalFilename = filepath.Join(s.LocalPath, file.Name())
			}
			// fmt.Printf("::: [%s] Season %d Episode %d [FOUND]\n", s.SeriesName, season, episode)
		}
		return nil
	})
	// fmt.Printf("Could not check directory for series (%s) %v\n", s.SeriesName, err)
}

func main() {
	app := cli.NewApp()

	//
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

		// Hold data for all series found locally
		var seriesList = make(map[string]Series)

		// Loop tvpath for folders and try to match them with series from TvDB
		files, err := ioutil.ReadDir(tvpath)
		if err != nil {
			fmt.Println("Could not list directory contents", err)
		} else {
			for _, folder := range files {
				if folder.IsDir() && !strings.HasPrefix(folder.Name(), ".") {
					// fmt.Println("Trying to find series (", folder.Name(), ")")
					// Try to match each series according to folder name
					series, err := getSeriesInfo(folder.Name())
					if err != nil {
						fmt.Println("Could not match series (", folder.Name(), ") with error ", err)
					} else {
						series.LocalPath = filepath.Join(tvpath, folder.Name())
						seriesList[folder.Name()] = series
						// Fill in which episodes exists locally
						series.CheckForExistingEpisodes()
					}
				}
			}
		}
		for _, series := range seriesList {
			fmt.Printf("Series '%s'\n", series.SeriesName)
			for _, episode := range series.Episodes {
				fmt.Printf(" > Season %d Episode %d ", episode.SeasonNumber, episode.EpisodeNumber)
				if episode.HasAired {
					if episode.LocalExists {
						fmt.Printf("is available locally\n")
					} else {
						fmt.Printf("is unavailable locally\n")
					}
				} else {
					fmt.Printf("has not aired yet\n")
				}
			}
		}
	}
	app.Run(os.Args)
}

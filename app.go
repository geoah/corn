package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/garfunkel/go-tvdb"
)

// GetSeries from tvdbcom
func getShowInfo(seriesName string) (tvdb.Series, error) {
	seriesList, err := tvdb.GetSeries(seriesName)

	if err != nil {
		return tvdb.Series{}, err
	}

	if len(seriesList.Series) > 0 {
		var series = *seriesList.Series[0]
		series.GetDetail()
		return series, nil
	} else {
		return tvdb.Series{}, errors.New("Not found")
	}
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
		tvpath := c.Args()[0]

		// Open tvpath directory
		dir, err := os.Open(tvpath)
		if err != nil {
			return
		}
		defer dir.Close()

		// Loop all directories
		// TODO Errors
		fileInfos, err := dir.Readdir(-1)
		if err != nil {
			return
		}
		for _, fi := range fileInfos {
			// TODO Check if is folder
			// TODO Ignore any file starting with a special charachter
			fmt.Println("Trying to find series (", fi.Name(), ")")
			// Try to find each series according to folder name
			series, err := getShowInfo(fi.Name())
			if err == nil {
				for _, seasonEpisodes := range series.Seasons {
					for _, episode := range seasonEpisodes {
						aired, err := time.Parse("2006-02-01", episode.FirstAired)
						if err != nil {

						} else {
							if aired.Before(time.Now()) {
								fmt.Println(series.SeriesName, "Season", episode.SeasonNumber, "Episode", episode.EpisodeNumber, "aired", episode.FirstAired)
							} else {
								fmt.Println(series.SeriesName, "Season", episode.SeasonNumber, "Episode", episode.EpisodeNumber, "not yet aired, airing on", episode.FirstAired)
							}
						}
					}
				}
			} else {
				fmt.Println("Could not match series with error ", err)
			}
		}
	}
	app.Run(os.Args)
}

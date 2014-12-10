package main

import (
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/garfunkel/go-tvdb"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
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

func checkSeries(series tvdb.Series) {
	for _, seasonEpisodes := range series.Seasons {
		for _, episode := range seasonEpisodes {
			if episode.FirstAired == "" {
				// fmt.Println("Missing first aired.")
			} else {
				aired, err := time.Parse("2006-01-02", episode.FirstAired)
				if err != nil {
					fmt.Println("Could not parse first aired.", err)
				} else {
					if aired.Before(time.Now()) {
						fmt.Println(series.SeriesName, "Season", episode.SeasonNumber, "Episode", episode.EpisodeNumber, "aired", episode.FirstAired)
					} else {
						fmt.Println(series.SeriesName, "Season", episode.SeasonNumber, "Episode", episode.EpisodeNumber, "not yet aired, airing on", episode.FirstAired)
					}
				}
			}
		}
	}
}

func getExistingEpisodes(seriesPath string) {
	regOne := regexp.MustCompile("[Ss]([0-9]+)[][ ._-]*[Ee]([0-9]+)([^\\/]*)$")

	files, err := ioutil.ReadDir(seriesPath)
	if err == nil {
		for _, file := range files {
			if !file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
				res := regOne.FindAllStringSubmatch(file.Name(), -1)
				season := res[0][1]
				episode := res[0][2]
				fmt.Println("Season:", season, "Episode:", episode)
			}
		}

	} else {
		fmt.Println(err)
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

		files, err := ioutil.ReadDir(tvpath)
		if err != nil {
			fmt.Println(err)
		} else {
			for _, folder := range files {
				if folder.IsDir() && !strings.HasPrefix(folder.Name(), ".") {
					// fmt.Println("Trying to find series (", folder.Name(), ")")
					// Try to find each series according to folder name
					series, err := getShowInfo(folder.Name())
					if err == nil {
						go checkSeries(series)
					} else {
						fmt.Println("Could not match series (", folder.Name(), ") with error ", err)
					}
				}
			}
		}
	}
	app.Run(os.Args)
}

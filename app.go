package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/codegangsta/cli"
	"github.com/jackpal/Taipei-Torrent/torrent"
	"github.com/kennygrant/sanitize"
)

// getShow retrieves a tv show from eztvapi using its imdb id and returns the a Show, or nil
func getShow(imdbid string) Show {
	url := "http://eztvapi.re/show/" + imdbid
	res, err := http.Get(url)
	// TODO Check for errors
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	// TODO Check for errors
	var data Show
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Somethign went wrong.")
		fmt.Printf("%T\n%s\n%#v\n", err, err, err)
		switch v := err.(type) {
		case *json.SyntaxError:
			fmt.Println(string(body[v.Offset-40 : v.Offset]))
		}
	} else {
		fmt.Println(data.Title)
	}
	return data
}

func downloadTorrentFromMagnetLinks(magnetLinks []string) {
	var torrentFlags torrent.TorrentFlags
	torrentFlags = torrent.TorrentFlags{
		//Dial:                dialerFromFlags(),
		//Port:                *port,
		//FileDir:             *fileDir,
		//SeedRatio:           *seedRatio,
		//UseDeadlockDetector: *useDeadlockDetector,
		//UseLPD:              *useLPD,
		UseDHT: true,
		//UseUPnP:             *useUPnP,
		//UseNATPMP:           *useNATPMP,
		//TrackerlessMode:     *trackerlessMode,
		// IP address of gateway
		//Gateway: *gateway,
	}

	log.Println("Starting.")

	err := torrent.RunTorrents(&torrentFlags, magnetLinks)
	if err != nil {
		log.Fatal("Could not run torrents", magnetLinks, err)
	}
}

func main() {
	app := cli.NewApp()

	//
	app.Action = func(c *cli.Context) {
		imdbid := c.Args()[0]
		tvpath := c.Args()[1]

		// TODO Check for first arg
		fmt.Println("Trying to get show with imdb id: ", imdbid)
		fmt.Println("> Available episodes: ")

		// Get the Show
		var show Show = getShow(imdbid)

		// Magnet Links
		magnetLinks := make([]string, 0, len(show.Episodes))

		// For each episode get available resolutions and check if they exist on the given directory (tvpath)
		for _, episode := range show.Episodes {
			fmt.Printf("> > S%dE%d", int(episode.Season), int(episode.Episode))
			if episode.Torrents.Hd720p.URL != "" {
				fmt.Printf(" @ 720p")
			} else if episode.Torrents.Sd480p.URL != "" {
				fmt.Printf(" @ 480p")
			} else {
				fmt.Printf(" @ sdtv")

			}
			magnetLinks = append(magnetLinks, episode.Torrents.Sd.URL)
			var episodepath string = fmt.Sprintf("%s/%s/season.%d", tvpath, sanitize.Path(strings.Replace(show.Title, " ", ".", -1)), int(episode.Season))
			fmt.Printf(" will be stored under '%s/'", episodepath)
			fmt.Printf("\n")
		}
		fmt.Println("")
		log.Println(magnetLinks)
		downloadTorrentFromMagnetLinks(magnetLinks)
	}
	app.Run(os.Args)
}

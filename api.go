package main

import (
	"fmt"
	"github.com/codegangsta/martini-contrib/encoder"
	"github.com/go-martini/martini"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

func PopSeries() {
	// Get tvpath from args
	tvpath := config.tvPath

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
		fmt.Printf("Going through %s\n", tvpath)
		for _, folder := range files {
			if folder.IsDir() && !strings.HasPrefix(folder.Name(), ".") {
				// Add to queue
				wg.Add(1)
				go func(tvpath string, folderName string) {
					var series Series = Series{}
					series.LocalName = folderName
					series.LocalPath = filepath.Join(tvpath, folderName)
					series.fetchInfo()
					// TODO Retry, Sometimes the tv api fucks up
					if series.Matched == true {
						store.Add(&series)
						fmt.Printf("Added series '%s' as '%s' to store.\n", series.LocalName, series.SeriesName)
					} else {
						fmt.Printf("[error] Could not add series '%s'.\n", series.LocalName, series.SeriesName)
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

func AddEpisode(enc encoder.Encoder, store Store, parms martini.Params) (int, []byte) {
	// TODO Check for parms:id
	// Get payload Object from Store
	id, err := strconv.ParseUint(parms["id"], 10, 64)
	// eid, err := strconv.ParseUint(parms["eid"], 10, 64)
	eid := parms["eid"]
	fmt.Println(id, eid)
	if err != nil {
		return http.StatusBadRequest, encoder.Must(enc.Encode(err))
	} else {
		series := store.Get(id)
		series.fetchDetails()
		if series.Matched == true {
			series.CheckForExistingEpisodes()
			series.FetchTorrentLinks()
			// series.PrintResults()
			// series.PrintJsonResults()
			episode := store.GetEpisode(id, eid)
			episode.start()
			return http.StatusOK, encoder.Must(enc.Encode(episode))
		}
		// TODO Check if payload exists
		return http.StatusNotFound, nil
	}
}

func GetEpisode(enc encoder.Encoder, store Store, parms martini.Params) (int, []byte) {
	// TODO Check for parms:id
	// Get payload Object from Store
	id, err := strconv.ParseUint(parms["id"], 10, 64)
	// eid, err := strconv.ParseUint(parms["eid"], 10, 64)
	eid := parms["eid"]
	fmt.Println(id, eid)
	if err != nil {
		return http.StatusBadRequest, encoder.Must(enc.Encode(err))
	} else {
		series := store.Get(id)
		series.fetchDetails()
		if series.Matched == true {
			series.CheckForExistingEpisodes()
			series.FetchTorrentLinks()
			// series.PrintResults()
			// series.PrintJsonResults()
			episode := store.GetEpisode(id, eid)
			return http.StatusOK, encoder.Must(enc.Encode(episode))
		}
		// TODO Check if payload exists
		return http.StatusNotFound, nil
	}

}

func GetAllSeries(r *http.Request, enc encoder.Encoder, store Store) []byte {
	return encoder.Must(enc.Encode(store.GetAll()))
}

func GetSeries(enc encoder.Encoder, store Store, parms martini.Params) (int, []byte) {
	// TODO Check for parms:id
	// Get payload Object from Store
	id, err := strconv.ParseUint(parms["id"], 10, 64)
	if err != nil {
		return http.StatusBadRequest, encoder.Must(enc.Encode(err))
	} else {
		series := store.Get(id)
		series.fetchDetails()
		if series.Matched == true {
			series.CheckForExistingEpisodes()
			series.FetchTorrentLinks()
			// series.PrintResults()
			// series.PrintJsonResults()
		}
		// TODO Check if payload exists
		return http.StatusOK, encoder.Must(enc.Encode(series))
	}
}

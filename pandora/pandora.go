package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/stayradiated/grooveshark"
	gs "github.com/stayradiated/grooveshark/responses"
	"github.com/stayradiated/pandora"
)

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func findMatch(song pandora.Song, tracks []gs.Track) gs.Track {
	return tracks[0]
}

func sanitize(name string) string {
	return strings.Replace(name, "/", "_", -1)
}

func download(id int, client *grooveshark.Client, songs <-chan pandora.Song, results chan<- bool) {
	for song := range songs {
		query := song.Name + " " + song.Artist
		filename := "cache/" + sanitize(song.Name+" - "+song.Artist) + ".mp3"

		fmt.Println(id, "##", query)

		// don't overwrite files
		if exists, _ := exists(filename); exists {
			fmt.Println(id, "> File exists")
			results <- false
			continue
		}

		fmt.Println(id, "> Searching grooveshark for match")

		// search grooveshark for the track
		tracks := client.Search(query)
		if len(tracks) < 1 {
			fmt.Println(id, "> Couldn't find a match")
			results <- false
			continue
		}

		// just take the first match
		// this could be improved to look for higher quality tracks
		track := findMatch(song, tracks)

		// download the song from grooveshark
		streamKey, err := client.GetStreamKey(track.SongId)
		if err != nil {
			fmt.Println(id, "> Couldn't get a streamKey")
			results <- false
			continue
		}

		resp, err := streamKey.Download()
		if err != nil {
			fmt.Println(id, "> Couldn't download song")
			results <- false
			continue
		}
		defer resp.Body.Close()

		// save the file to the cache
		output, err := os.Create(filename)
		if err != nil {
			panic(err)
		}
		defer output.Close()

		fmt.Println(id, "> Downloading")

		// Copy from the download stream to the file write stream
		_, err = io.Copy(output, resp.Body)
		if err != nil {
			panic(err)
		}

		// Let the pool know we have finished
		results <- true
	}
}

func main() {

	username := os.Args[1]
	password := os.Args[2]

	fmt.Println("Extracting tracks from Pandora")

	stations, err := pandora.FetchStations(username, password)
	if err != nil {
		panic(err)
	}

	fmt.Println("Have stations:", len(stations))

	client := grooveshark.NewClient()
	client.Connect()

	fmt.Println("Connected to Grooveshark")

	total := 0
	songs := make(chan pandora.Song, 1000)
	results := make(chan bool, 1000)

	// start three workers
	for w := 0; w < 5; w++ {
		go download(w, client, songs, results)
	}

	// pass songs into workers
	for _, station := range stations {
		for _, song := range station.Songs {
			total += 1
			songs <- song
		}
	}
	close(songs)

	// wait for results
	for i := 0; i < total; i++ {
		<-results
	}

}

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/stayradiated/grooveshark"
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

func download(client *grooveshark.Client, song pandora.Song) {
	query := song.Name + " " + song.Artist
	filename := "cache/" + song.Name + " - " + song.Artist + ".mp3"

	if exists, _ := exists(filename); exists {
		return
	}

	fmt.Println("Seaching", query)

	tracks := client.Search(query)
	if len(tracks) < 1 {
		fmt.Println("> Couldn't find a match")
		return
	}

	track := tracks[0]

	streamKey := client.GetStreamKey(track.SongId)

	resp, err := streamKey.Download()
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	output, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer output.Close()

	fmt.Println("Downloading", filename)

	_, err = io.Copy(output, resp.Body)
	if err != nil {
		panic(err)
	}
}

func main() {

	username := os.Args[1]
	password := os.Args[2]

	stations, err := pandora.FetchStations(username, password)
	if err != nil {
		panic(err)
	}

	client := grooveshark.NewClient()
	client.Connect()

	for _, station := range stations {
		for _, song := range station.Songs {
			download(client, song)
		}
	}

}

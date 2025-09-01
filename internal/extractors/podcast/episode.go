package podcast

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

type Episode struct {
	PodcastTitle string
	Number       int
	Title        string
	FileURL      string
	Date         string
}

// Get data from Feed and put it into an Episode struct,
// returns a slice of episodes.
func GetEpisodes(feed *gofeed.Feed) []Episode {
	items := feed.Items

	episodes := []Episode{}

	for _, entry := range items {
		var audioFileURL string
		for _, item := range entry.Enclosures {
			// TODO: Need to make sure it's an audio link
			audioFileURL = item.URL
		}
		episode := Episode{
			PodcastTitle: feed.Title,
			//	Number:       counter,
			Title:   entry.Title,
			FileURL: audioFileURL,
			Date:    entry.Published,
		}
		episodes = append(episodes, episode)
	}

	// Sort slice by date in descending order
	sort.Slice(episodes, func(a, b int) bool {
		time1, _ := time.Parse(time.RFC1123Z, episodes[a].Date)
		time2, _ := time.Parse(time.RFC1123Z, episodes[b].Date)
		return time1.After(time2)
	})

	counter := len(episodes)
	for index := range episodes {
		counter--
		episodes[index].Number = counter
	}

	return episodes
}

// Boolean check if filepath exists.
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)

	if err != nil {
		return false
	}
	return true
}

// Download audio file from an episode.
func DownloadFile(episode Episode) error {
	// Common character that some devices don't like.
	rep := strings.NewReplacer(
		" ", "_",
		":", "_",
		"(", "-",
		")", "-",
		"[", "-",
		"]", "-",
	)
	episode.Title = rep.Replace(episode.Title)
	fileName := fmt.Sprintf("%d._%s.mp3", episode.Number, episode.Title)
	homeDirectory, err := os.UserHomeDir()
	filePath := fmt.Sprintf("%s/Music/podcasts/%s/%s", homeDirectory, episode.PodcastTitle, fileName)

	if fileExists(filePath) {
		// TODO: Maybe ask if user wants to overwrite?
		return errors.New("File name already exists")
	}

	err = os.MkdirAll(filepath.Dir(filePath), 0770)

	if err != nil {
		return err
	}

	file, err := os.Create(filePath)

	if err != nil {
		fmt.Println("CANT CREATE FILE")
		return err
	}

	defer file.Close()

	resp, err := http.Get(episode.FileURL)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// Get feed from url, cancels if it takes 10+ seconds.
func GetPodcastFeed(url string) (*gofeed.Feed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a custom HTTP client with better TLS configuration
	transport := &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
		// Add retry logic for TLS issues
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	// Create parser with custom HTTP client
	feedparser := gofeed.NewParser()
	feedparser.Client = client

	feed, err := feedparser.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, err
	}

	return feed, nil
}

package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

func main() {
	start := time.Now()
	current := start

	titles := make(map[string]OuiFmTitle)

	timestamp := fmt.Sprintf("%d.%d.%dT%d.%d", start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute())
	f, err := os.Create(fmt.Sprintf("freeYourMusic.%s.csv", timestamp))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	cw := csv.NewWriter(f)

	for start.Sub(current).Hours() < 24*7 {
		u, _ := url.Parse("https://www.ouifm.fr/api/TitleDiffusions")
		q := u.Query()
		q.Set("size", "30") // max allowed
		q.Set("radioStreamId", "2174546520932614531")
		q.Set("date", strconv.FormatInt(current.UnixMilli(), 10))
		u.RawQuery = q.Encode()

		log.Printf("harvesting 30 titles @%v : %s", current, u.String())
		resp, err := http.DefaultClient.Get(u.String())
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode > 299 {
			log.Fatalln("error from server", resp.StatusCode, resp.Status, string(b))
		}

		parsedResponse := OuiFmResponse{}

		err = json.Unmarshal(b, &parsedResponse)
		if err != nil {
			log.Fatal(err)
		}

		for _, entry := range parsedResponse {
			titles[entry.ID] = entry.Title

			if entry.Timestamp.Before(current) {
				current = entry.Timestamp
			}
		}
	}

	for _, title := range titles {
		err = cw.Write([]string{title.Title, title.Artist})
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("harvested %d titles\n", len(titles))
}

type OuiFmResponse []struct {
	Typename  string     `json:"__typename"`
	ID        string     `json:"id"`
	Timestamp time.Time  `json:"timestamp"`
	MdsID     string     `json:"mdsId"`
	Title     OuiFmTitle `json:"title"`
}

type OuiFmTitle struct {
	Typename  string `json:"__typename"`
	ID        string `json:"id"`
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	CoverURL  string `json:"coverUrl"`
	SpotifyID string `json:"spotifyId"`
	DeezerID  string `json:"deezerId"`
	CoverID   string `json:"coverId"`
}

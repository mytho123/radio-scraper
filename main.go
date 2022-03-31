package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var (
	radioStreamIds = map[string]string{
		"ouifm":                 "2174546520932614531",
		"ouifm_rock_inde":       "3134161803443976526",
		"ouifm_bring_the_noise": "4004502594738215513",
		"ouifm_reggae":          "3540892623380233022",
		"ouifm_accoustic":       "3906034555622012146",
		"latina":                "2174546520932614634",
		"voltage":               "2174546520932614870",
	}
)

func main() {
	help := flag.Bool("help", false, "Displays this help")
	flag.Parse()

	if *help || len(flag.Args()) == 0 {
		const bold = "\033[1m"
		const reset = "\033[0m"

		os.Stdout = os.Stderr
		fmt.Println("Usage: rm [OPTION]... [SOURCES]...")
		fmt.Println("Scraps the given sources to a CSV file")
		fmt.Println()
		fmt.Println(bold, "Options", reset)
		flag.CommandLine.PrintDefaults()

		fmt.Println()
		fmt.Println(bold, "Available sources", reset)
		for k := range radioStreamIds {
			fmt.Println("\t" + k)
		}

		os.Exit(0)
	}

	for _, source := range flag.Args() {
		streamId, ok := radioStreamIds[source]
		if !ok {
			log.Fatalf("%s is not a valid source\n", source)
		}
		harvestSource(source, streamId)
	}
}

func harvestSource(name string, sourceId string) {
	start := time.Now()
	current := start

	titles := make(map[string]OuiFmTitle)

	timestamp := fmt.Sprintf("%d.%d.%dT%d.%d", start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute())
	f, err := os.Create(fmt.Sprintf("%s.%s.csv", name, timestamp))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	cw := csv.NewWriter(f)

	for start.Sub(current).Hours() < 24*7 {
		u, _ := url.Parse("https://www.ouifm.fr/api/TitleDiffusions")
		q := u.Query()
		q.Set("size", "30") // max allowed
		q.Set("radioStreamId", sourceId)
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

package sources

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func scrapGroupe1981(sourceID string, start, end time.Time, results chan<- Track) error {
	type response []struct {
		Typename  string    `json:"__typename"`
		ID        string    `json:"id"`
		Timestamp time.Time `json:"timestamp"`
		MdsID     string    `json:"mdsId"`
		Title     struct {
			Typename  string `json:"__typename"`
			ID        string `json:"id"`
			Title     string `json:"title"`
			Artist    string `json:"artist"`
			CoverURL  string `json:"coverUrl"`
			SpotifyID string `json:"spotifyId"`
			DeezerID  string `json:"deezerId"`
			CoverID   string `json:"coverId"`
		} `json:"title"`
	}

	current := end

	for start.Before(current) {
		u, _ := url.Parse("https://www.ouifm.fr/api/TitleDiffusions")
		q := u.Query()
		q.Set("size", "30") // max allowed
		q.Set("radioStreamId", sourceID)
		q.Set("date", strconv.FormatInt(current.UnixMilli(), 10))
		u.RawQuery = q.Encode()

		log.Printf("scrapping 30 titles @%v : %s", current, u.String())
		resp, err := http.DefaultClient.Get(u.String())
		if err != nil {
			return err
		}

		defer func() { _ = resp.Body.Close() }()

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode > 299 {
			log.Fatalln("error from server", resp.StatusCode, resp.Status, string(b))
		}

		parsedResponse := response{}
		err = json.Unmarshal(b, &parsedResponse)
		if err != nil {
			return err
		}

		for _, entry := range parsedResponse {
			results <- Track{
				Artist: entry.Title.Artist,
				Title:  entry.Title.Title,
			}

			if entry.Timestamp.Before(current) {
				current = entry.Timestamp
			}
		}
	}

	close(results)

	return nil
}

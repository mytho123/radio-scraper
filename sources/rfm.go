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

func scrapRfm(sourceURL string, start, end time.Time, results chan<- Track) error {
	type response struct {
		Nowplaying []struct {
			ID       int       `json:"id"`
			Artist   string    `json:"artist"`
			Title    string    `json:"title"`
			Time     time.Time `json:"time"`
			ImageURL string    `json:"imageUrl"`
			Duration string    `json:"duration"`
			Status   string    `json:"status"`
			Type     string    `json:"type"`
		} `json:"nowplaying"`
	}

	u, err := url.Parse(sourceURL)
	if err != nil {
		return err
	}

	current := end

	for start.Before(current) {
		q := u.Query()
		q.Set("limit", "20")
		q.Set("ts", strconv.FormatInt(current.Unix(), 10))
		u.RawQuery = q.Encode()

		log.Printf("scrapping 20 titles @%v : %s", current, u.String())
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

		for _, entry := range parsedResponse.Nowplaying {
			if entry.Artist == "" || entry.Title == "" {
				continue
			}

			results <- Track{
				Artist: entry.Artist,
				Title:  entry.Title,
			}

			if entry.Time.Before(current) {
				current = entry.Time
			}
		}
	}

	close(results)

	return nil
}

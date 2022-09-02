package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/mytho123/radio-scraper/sources"
	"hash/fnv"
	"log"
	"os"
	"time"
)

type Aggregator interface {
	GetSources() []string
	Scrap(source string, start, end time.Time, results chan<- sources.Track) error
}

func main() {
	help := flag.Bool("help", false, "Displays this help")
	duration := flag.String("d", "168h", "Duration to parse (defaults to 1 week)")
	flag.Parse()

	allSources := sources.GetSources()

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
		for name := range allSources {
			fmt.Println("\t" + name)
		}

		os.Exit(0)
	}

	durationParsed, err := time.ParseDuration(*duration)
	if err != nil {
		log.Fatal(err)
	}

	for _, source := range flag.Args() {
		scrap, ok := allSources[source]
		if !ok {
			log.Fatalf("%s is not a valid source\n", source)
		}
		harvestSource(source, durationParsed, scrap)
	}
}

func harvestSource(name string, duration time.Duration, scrap sources.Scrap) {
	start := time.Now()
	timestamp := fmt.Sprintf("%d.%d.%dT%d.%d", start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute())
	f, err := os.Create(fmt.Sprintf("%s.%s.csv", name, timestamp))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	results := make(chan sources.Track, 100)

	go func() {
		now := time.Now()
		err := scrap(now.Add(-duration), now, results)
		if err != nil {
			return
		}
	}()

	cw := csv.NewWriter(f)
	defer cw.Flush()
	total := 0
	duplicates := 0

	hashes := make(map[uint32]interface{})

	for track := range results {
		h := fnv.New32a()
		_, _ = h.Write([]byte(fmt.Sprintf("%s|%s", track.Title, track.Artist)))
		sum := h.Sum32()

		_, dupl := hashes[sum]
		if dupl {
			duplicates++
			continue
		}
		hashes[sum] = nil

		err = cw.Write([]string{track.Title, track.Artist})
		if err != nil {
			log.Fatal(err)
		}

		total++
	}

	fmt.Printf("harvested %d titles (%d duplicates)\n", total, duplicates)
}

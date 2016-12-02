package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	scrapeGames     = flag.Bool("scrape", false, "Scrape the latest list of games")
	fullScrapeGames = flag.Bool("fullscrape", false, "Scrape and do not stop on duplicates")
	updateGames     = flag.Int("update", -1, "Update the list of games not updated within the last N minutes")
	listGames       = flag.Int("list", 0, "List top N games")
	dbPath          = flag.String("database", "", "Database location (default $HOME/.config/steamdb/steam.db)")
)

func getDatabaseFile() string {
	if *dbPath != "" {
		return *dbPath
	}
	dir := filepath.Join(os.Getenv("HOME"), ".config", "steamdb")
	if err := os.MkdirAll(dir, 0700); err != nil {
		panic(fmt.Errorf("Unable to make database directory: %v", err))
	}
	return filepath.Join(dir, "steam.db")
}

func main() {
	flag.Parse()
	if !*scrapeGames && !*fullScrapeGames && *updateGames < 0 && *listGames < 1 {
		flag.Usage()
		os.Exit(1)
	}
	gamesDb, err := NewGamesDb(getDatabaseFile())
	if err != nil {
		panic(fmt.Errorf("error connecting to games database: %v", err))
	}
	defer gamesDb.Close()
	if *scrapeGames || *fullScrapeGames {
		if err := NewScraper(gamesDb).Scrape(*fullScrapeGames); err != nil {
			log.Fatalf("Error during scrape: %v", err)
		}
	}
	if *updateGames >= 0 {
		if err := NewUpdater(gamesDb).Update(time.Duration(*updateGames) * time.Minute); err != nil {
			log.Fatalf("Error during update: %v", err)
		}
	}
	if *listGames > 0 {
		if err := NewLister(gamesDb).List(*listGames); err != nil {
			log.Fatalf("Error during list: %v", err)
		}
	}

}

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/explodes/ezconfig"
	_ "github.com/explodes/ezconfig/db/pg"
)

const (
	maxJsonGames = 10000
)

var (
	scrapeGames     = flag.Bool("scrape", false, "Scrape the latest list of games")
	fullScrapeGames = flag.Bool("fullscrape", false, "Scrape and do not stop on duplicates")
	updateGames     = flag.Int("update", -1, "Update the list of games not updated within the last N minutes")
	listGames       = flag.Int("list", 0, "List top N games")
	config          = flag.String("config", "", "Database configuration file")
)

var (
	serve             = flag.Bool("serve", false, "Run as a web service")
	serveUpdatePeriod = flag.Int("update-period", 5, "Update stats every N minutes")
	serveGamesPeriod  = flag.Int("games-period", 60, "Update games every N minutes")
	serveAddr         = flag.String("port", ":9654", "Server bind address")
)

func getDatabaseConfig() *ezconfig.DbConfig {
	if *config == "" {
		panic(errors.New("Config file not specified"))
	}
	conf := &ezconfig.DbConfig{}
	if err := ezconfig.ReadConfig(*config, conf); err != nil {
		panic(err)
	}
	return conf
}

func main() {
	flag.Parse()

	if !*serve && !*scrapeGames && !*fullScrapeGames && *updateGames < 0 && *listGames < 1 {
		flag.Usage()
		os.Exit(1)
	}
	gamesDb, err := NewGamesDb(getDatabaseConfig())
	if err != nil {
		panic(fmt.Errorf("error connecting to games database: %v", err))
	}

	if *serve {
		runGamesServer(gamesDb)
		return
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

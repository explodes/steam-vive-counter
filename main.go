package main

import (
	"flag"
	"fmt"
	"github.com/explodes/jsonserv"
	"log"
	"net/http"
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

var (
	serve             = flag.Bool("serve", false, "Run as a web service")
	serveUpdatePeriod = flag.Int("update-period", 5, "Update stats every N minutes")
	serveGamesPeriod  = flag.Int("games-period", 60, "Update games every N minutes")
	serveAddr         = flag.String("port", ":9654", "Server bind address")
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

	if !*serve && !*scrapeGames && !*fullScrapeGames && *updateGames < 0 && *listGames < 1 {
		flag.Usage()
		os.Exit(1)
	}
	gamesDb, err := NewGamesDb(getDatabaseFile())
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

type gameJson struct {
	AppId   int64  `json:"app_id"`
	Name    string `json:"name"`
	Players int    `json:"players"`
	Rank    int    `json:"rank"`
}

type serverContext struct {
	games []gameJson
}

func runGamesServer(db *GamesDb) {

	update := time.Tick(time.Duration(*serveUpdatePeriod) * time.Minute)
	games := time.Tick(time.Duration(*serveGamesPeriod) * time.Minute)

	ctx := &serverContext{
		games: make([]gameJson, 0),
	}

	updateServerGamesList(db, ctx)

	go func() {
		scraper := NewScraper(db)
		updater := NewUpdater(db)
		for {
			select {
			case <-update:
				if err := updater.Update(time.Duration(*serveUpdatePeriod)); err != nil {
					log.Printf("error updating: %v", err)
				}
			case <-games:
				if err := scraper.Scrape(false); err != nil {
					log.Printf("error updating: %v", err)
				}
				if err := updateServerGamesList(db, ctx); err != nil {
					log.Printf("error building games list: %v", err)
				}
			}
		}
	}()

	log.Printf("Serving at %s", *serveAddr)
	err := jsonserv.New().
		SetContext(ctx).
		AddMiddleware(jsonserv.NewLoggingMiddleware(false)).
		AddMiddleware(jsonserv.NewMaxRequestSizeMiddleware(128)).
		AddMiddleware(jsonserv.NewDebugFlagMiddleware(false)).
		AddMiddleware(jsonserv.NewGzipMiddleware()).
		AddRoute(http.MethodGet, "games", "/", gamesView).
		Serve(*serveAddr)
	if err != nil {
		log.Fatal(err)
	}

}

func updateServerGamesList(db *GamesDb, ctx *serverContext) error {
	results := make([]gameJson, 0)

	iter, err := db.GetTopGames(1000)
	if err != nil {
		return err
	}
	defer iter.Close()
	for i := 0; iter.Next(); i++ {
		game, err := iter.Game()
		if err != nil {
			return err
		}
		results = append(results, gameJson{
			AppId:   game.AppId,
			Name:    game.Name,
			Players: game.Players,
			Rank:    i + 1,
		})
	}
	ctx.games = results
	return nil
}

func gamesView(c interface{}, req *jsonserv.Request, res *jsonserv.Response) {
	ctx := c.(*serverContext)
	res.Ok(ctx.games)
}

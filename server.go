package main

import (
	"log"
	"net/http"
	"time"

	"github.com/explodes/jsonserv"
)

type gameJson struct {
	AppId   int64  `json:"app_id"`
	Name    string `json:"name"`
	Players int    `json:"players"`
	Rank    int    `json:"rank"`
}

type serverContext struct {
	games []gameJson
}

func ticker(minutes int) *time.Ticker {
	return time.NewTicker(time.Duration(minutes) * time.Minute)
}

func runGamesServer(db *GamesDb) {

	update := ticker(*serveUpdatePeriod)
	games := ticker(*serveGamesPeriod)

	ctx := &serverContext{
		games: make([]gameJson, 0),
	}

	updateServerGamesList(db, ctx)

	go func() {
		defer update.Stop()
		defer games.Stop()
		scraper := NewScraper(db)
		updater := NewUpdater(db)
		updateTheGames := func() {
			log.Print("updating")
			if err := updater.Update(time.Duration(*serveUpdatePeriod)); err != nil {
				log.Printf("error updating: %v", err)
			}
		}
		updateTheGames()
		for {
			select {
			case <-update.C:
				updateTheGames()
			case <-games.C:
				log.Print("scraping games")
				if err := scraper.Scrape(true); err != nil {
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
		SetApp(ctx).
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

	iter, err := db.GetTopGames(maxJsonGames)
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

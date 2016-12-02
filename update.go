package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	urlAppPlayers = `https://api.steampowered.com/ISteamUserStats/GetNumberOfCurrentPlayers/v1/?appid=%d`
)

type Updater struct {
	db     *GamesDb
	client *Client
	ctx    context.Context
	cancel context.CancelFunc
	err    error
}

func NewUpdater(db *GamesDb) *Updater {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &Updater{
		db:     db,
		client: NewClient(),
		ctx:    ctx,
		cancel: cancelFunc,
	}
}

func (u *Updater) Update(staleness time.Duration) error {
	since := time.Now().Add(-staleness)
	staleIds, err := u.db.GetUnUpdatedAppIds(since)
	if err != nil {
		return err
	}
	wg := &sync.WaitGroup{}
	for _, id := range staleIds {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			u.updateApp(id)
		}(id)
	}
	wg.Wait()
	return u.err
}

func (u *Updater) Cancel(err error) {
	u.err = err
	u.cancel()
}

func (u *Updater) updateApp(id int64) {
	select {
	case <-u.ctx.Done():
		return
	default:
		players := &NumberOfPlayers{}
		url := fmt.Sprintf(urlAppPlayers, id)
		if err := u.client.fetchJson(url, players); err != nil {
			u.Cancel(fmt.Errorf("error pulling app stats: %v", err))
			return
		}
		if err := u.db.UpdatePlayersCount(id, players.Response.PlayerCount); err != nil {
			u.Cancel(fmt.Errorf("error updating app %d stats: %v", id, err))
			return
		}
		log.Printf("Game %d has %d players", id, players.Response.PlayerCount)
	}
}

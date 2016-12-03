package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"sync"
)

const (
	searchUrl  = `http://store.steampowered.com/search/?sort_by=Released_DESC&vrsupport=101&page=%d`
	appInfoUrl = `http://store.steampowered.com/api/appdetails?appids=%d`
)

var appPageRegex = regexp.MustCompile(`/steam/apps/(\d+)`)

type Scraper struct {
	db     *GamesDb
	client *Client
	ctx    context.Context
	cancel context.CancelFunc
	err    error
}

func NewScraper(db *GamesDb) *Scraper {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scraper{
		db:     db,
		client: NewClient(),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Scraper) getAppIds(ids chan<- int64) {
	defer func() {
		ids <- -1
	}()
	for page := 1; ; page++ {
		select {
		case <-s.ctx.Done():
			return
		default:
			nextUrl := fmt.Sprintf(searchUrl, page)
			contents, err := s.client.fetchHtml(nextUrl)
			if err != nil {
				s.Cancel(fmt.Errorf("error fetching %s: %v", nextUrl, err))
				return
			}
			matches := appPageRegex.FindAllStringSubmatch(contents, -1)
			if matches == nil {
				return
			}
			for _, match := range matches {
				idString := match[1]
				id, err := strconv.ParseInt(idString, 10, 64)
				if err != nil {
					log.Printf("Bad app id: %s", idString)
					continue
				}
				ids <- id
			}
		}
	}
}

func (s *Scraper) saveAppInfo(id int64, continueOnDuplicate bool) {
	exists, err := s.db.Exists(id)
	if err != nil {
		s.Cancel(fmt.Errorf("error testing app %d: %v", id, err))
		return
	}
	if exists {
		if !continueOnDuplicate {
			s.Cancel(nil)
		}
		return
	}
	url := fmt.Sprintf(appInfoUrl, id)
	appInfoById := AppInfoById{}
	err = s.client.fetchJson(url, &appInfoById)
	if err != nil {
		log.Printf("error parsing json at %s: %v", url, err)
		return
	}
	if appInfoById == nil {
		s.Cancel(fmt.Errorf("empty json: %d, probably rate limit", id))
		return
	}
	stringId := fmt.Sprintf("%d", id)
	appInfo, ok := appInfoById[stringId]
	if !ok {
		s.Cancel(fmt.Errorf("unexpected app json: %d: %#v", id, appInfoById))
		return
	}
	dbId, err := s.db.SaveAppInfo(
		id,
		appInfo.Data.Name,
		appInfo.IsSingleplayer(),
		appInfo.IsMultiplayer(),
		appInfo.IsOnlineMultiplayer(),
		appInfo.IsLocalMultiplayer(),
	)
	if err != nil {
		s.Cancel(fmt.Errorf("error saving %s (%d): %v", appInfo.Data.Name, id, err))
		return
	}
	log.Printf("Saved %s (%d) to database: %d", appInfo.Data.Name, id, dbId)

}

func (s *Scraper) Cancel(err error) {
	s.err = err
	s.cancel()
}

func (s *Scraper) Scrape(continueOnDuplicate bool) error {
	ids := make(chan int64)
	go s.getAppIds(ids)
	wg := sync.WaitGroup{}
	for id := range ids {
		if id == -1 {
			break
		}
		select {
		case <-s.ctx.Done():
			break
		default:
			wg.Add(1)
			go func(id int64) {
				defer wg.Done()
				s.saveAppInfo(id, continueOnDuplicate)
			}(id)
		}
	}
	wg.Wait()
	return s.err
}

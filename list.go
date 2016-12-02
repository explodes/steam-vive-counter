package main

import (
	"fmt"
)

type Lister struct {
	db *GamesDb
}

func NewLister(db *GamesDb) *Lister {
	return &Lister{
		db: db,
	}
}

func (x *Lister) List(top int) error {
	iter, err := x.db.GetTopGames(top)
	if err != nil {
		return err
	}
	defer iter.Close()
	for i := 0; iter.Next(); i++ {
		game, err := iter.Game()
		if err != nil {
			return err
		}
		x.printGame(i+1, game)
	}
	return nil
}

func (x *Lister) printGame(rank int, game *Game) {
	fmt.Printf("%3d: %-6d %-35s %d\n", rank, game.AppId, game.Name, game.Players)
}

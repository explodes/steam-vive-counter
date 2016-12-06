package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/explodes/ezconfig"
	"github.com/explodes/ezconfig/opener"
)

const (
	sqlCreateGamesTable = `CREATE TABLE IF NOT EXISTS games (
	id BIGSERIAL PRIMARY KEY,
	app_id INTEGER NOT NULL,
	name TEXT NOT NULL,
	singleplayer INTEGER DEFAULT 0 NOT NULL,
	multiplayer INTEGER DEFAULT 0 NOT NULL,
	online_multiplayer INTEGER DEFAULT 0 NOT NULL,
	local_multiplayer INTEGER DEFAULT 0 NOT NULL,
	last_update INTEGER DEFAULT 0 NOT NULL,
	players INTEGER DEFAULT 0 NOT NULL
)`
	sqlGameInsert        = `INSERT INTO GAMES (app_id, name, singleplayer, multiplayer, online_multiplayer, local_multiplayer) VALUES ($1, $2, $3, $4, $5, $6) RETURNING ID`
	sqlGameExists        = `SELECT COUNT(id) FROM games WHERE app_id = $1`
	sqlGameUpdatePlayers = `UPDATE games SET players = $1, last_update = $2 WHERE app_id = $3`
	sqlGamesUnUpdated    = `SELECT app_id FROM games WHERE last_update < $1`
	sqlTopGames          = `SELECT id, app_id, name, singleplayer, multiplayer, online_multiplayer, local_multiplayer, last_update, players FROM games ORDER BY players DESC, name ASC LIMIT $1`
)

type Game struct {
	ID                int64
	AppId             int64
	Name              string
	Singleplayer      bool
	Multiplayer       bool
	OnlineMultiplayer bool
	LocalMultiplayer  bool
	LastUpdate        time.Time
	Players           int
}

type GamesDb struct {
	db *sql.DB
}

type GamesIter struct {
	game Game
	rows *sql.Rows
}

func connectDb(conf *ezconfig.DbConfig) (*sql.DB, error) {
	conn, err := opener.New().WithDatabase(conf).Connect()
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to database: %v", err)
	}
	if conn.DB == nil {
		return nil, fmt.Errorf("Unexpected nil database: %v", err)
	}
	return conn.DB, nil
}

func migrateDb(db *sql.DB, create string) error {
	if _, err := db.Exec(create); err != nil {
		return fmt.Errorf("Unable to migrate database: %v", err)
	}
	return nil
}

func NewGamesDb(config *ezconfig.DbConfig) (*GamesDb, error) {
	if config.Database.Type != "postgres" {
		return nil, fmt.Errorf("database not supported: %s", config.Database.Type)
	}
	conn, err := connectDb(config)
	if err != nil {
		return nil, err
	}
	if err = migrateDb(conn, sqlCreateGamesTable); err != nil {
		conn.Close()
		return nil, err
	}
	gamesDb := &GamesDb{
		db: conn,
	}
	return gamesDb, nil
}

func (g *GamesDb) Close() {
	g.db.Close()
}

func (g *GamesDb) Exists(id int64) (bool, error) {
	var count int
	if err := g.db.QueryRow(sqlGameExists, id).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (g *GamesDb) SaveAppInfo(appId int64, name string, singleplayer, multiplayer, onlineMultiplayer, localMultiplayer bool) (int64, error) {
	var id int64
	if err := g.db.QueryRow(sqlGameInsert, appId, name, b2i(singleplayer), b2i(multiplayer), b2i(onlineMultiplayer), b2i(localMultiplayer)).Scan(&id); err != nil {
		return -1, err
	}
	return id, nil
}

func (g *GamesDb) GetUnUpdatedAppIds(since time.Time) ([]int64, error) {
	timestamp := t2i(since)
	rows, err := g.db.Query(sqlGamesUnUpdated, timestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	appIds := make([]int64, 0)
	for rows.Next() {
		var appId int64
		if err := rows.Scan(&appId); err != nil {
			return nil, err
		}
		appIds = append(appIds, appId)
	}
	return appIds, nil
}

func (g *GamesDb) UpdatePlayersCount(id int64, players int) error {
	timestamp := t2i(time.Now())
	_, err := g.db.Exec(sqlGameUpdatePlayers, players, timestamp, id)
	return err
}

func (g *GamesDb) GetTopGames(limit int) (*GamesIter, error) {
	rows, err := g.db.Query(sqlTopGames, limit)
	if err != nil {
		return nil, err
	}
	iter := &GamesIter{
		rows: rows,
	}
	return iter, nil
}

func (g *GamesIter) Next() bool {
	return g.rows.Next()
}

func (g *GamesIter) Game() (*Game, error) {
	game := &g.game
	var update int64
	err := g.rows.Scan(&game.ID, &game.AppId, &game.Name, &game.Singleplayer, &game.Multiplayer, &game.OnlineMultiplayer, &game.LocalMultiplayer, &update, &game.Players)
	if err != nil {
		return nil, err
	}
	game.LastUpdate = i2t(update)
	return game, nil
}

func (g *GamesIter) Close() error {
	return g.rows.Close()
}

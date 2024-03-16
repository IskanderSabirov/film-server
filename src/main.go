package main

import (
	_ "embed"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"gopkg.in/yaml.v2"
	"time"
)

const (
	minimalFilmName        int = 1
	maximumFilmName        int = 150
	minimalFilmDescription int = 0
	maximumFilmDescription int = 1000
	minimalFilmRating      int = 0
	maximumFilmRating      int = 10

	filmsTable       string = "films"
	actorsTable      string = "actors"
	filmsActorsTable string = "films_actors"

	sortByName         string = "name"
	sortByPresentation string = "presentation"
)

var (
	dataBase LocalStorage
)

type Actor struct {
	name string    `json:"name"`
	sex  bool      `json:"sex"`
	born time.Time `json:"born"`
}

type Film struct {
	name         string    `json:"name"`
	description  string    `json:"description"`
	presentation time.Time `json:"presentation"`
	rating       int       `json:"rating"`
}

//go:embed db_config.yml
var rawDBConfig []byte

func main() {
	var dbConfig DBConfig
	var err error
	if err := yaml.Unmarshal(rawDBConfig, &dbConfig); err != nil {
		panic(err)
	}
	dataBase, err = NewDatabase(dbConfig,
		TablesNames{Films: filmsTable, Actors: actorsTable, FilmsActors: filmsActorsTable},
	)
	if err != nil {
		fmt.Print(err.Error())
	}
}

func checkFilmName(filmName string) bool {
	return len(filmName) >= minimalFilmName && len(filmName) <= maximumFilmName
}

func checkFilmRating(rating int) bool {
	return rating >= minimalFilmRating && rating <= maximumFilmRating
}

func checkFilmDescription(description string) bool {
	return len(description) >= minimalFilmDescription && len(description) <= maximumFilmDescription
}

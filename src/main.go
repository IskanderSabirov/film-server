package main

import (
	_ "embed"
	_ "github.com/jackc/pgx/v4/stdlib"
	"gopkg.in/yaml.v2"
	"log"
	"net/http"
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
	usersTable       string = "users"
	adminsTable      string = "admins"

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

type changedActor struct {
	prevName    string    `json:"prevName"`
	prevBorn    time.Time `json:"prevBorn"`
	prevSex     bool      `json:"prevSex"`
	nameChanged bool      `json:"nameChanged"`
	newName     string    `json:"newName"`
	sexChanged  bool      `json:"sexChanged"`
	newSex      bool      `json:"newSex"`
	bornChanged bool      `json:"bornChanged"`
	newBorn     time.Time `json:"newBorn"`
}

type changedFilm struct {
	prevName            string    `json:"prevName"`
	prevPresentation    time.Time `json:"prevPresentation"`
	nameChanged         bool      `json:"nameChanged"`
	newName             string    `json:"newName"`
	descriptionChanged  bool      `json:"descriptionChanged"`
	newDescription      bool      `json:"newDescription"`
	presentationChanged bool      `json:"presentationChanged"`
	newPresentation     time.Time `json:"newPresentation"`
	ratingChanged       bool      `json:"ratingChanged"`
	newRating           int       `json:"newRating"`
}

type sortFilms struct {
	sort string `json:"sort"`
}

type findSubstring struct {
	substring string `json:"substring"`
}
type addActorsToFilm struct {
	film   Film    `json:"film"`
	actors []Actor `json:"actors"`
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
		TablesNames{
			Films:       filmsTable,
			Actors:      actorsTable,
			FilmsActors: filmsActorsTable,
			Users:       usersTable,
			Admins:      adminsTable,
		},
	)
	if err != nil {
		panic(err)
	}

	//admin
	http.HandleFunc("/admin/addActor", adminAuthenticate(addActorHandler))
	http.HandleFunc("/admin/changeActor", adminAuthenticate(changeActorHandler))
	http.HandleFunc("/admin/deleteActor", adminAuthenticate(deleteActorHandler))

	http.HandleFunc("/admin/addFilm", adminAuthenticate(addFilmHandler))
	http.HandleFunc("/admin/changeFilm", adminAuthenticate(changeFilmHandler))
	http.HandleFunc("/admin/deleteFilm", adminAuthenticate(deleteFilmHandler))

	http.HandleFunc("/admin/getFilms", adminAuthenticate(getFilmsHandler))
	http.HandleFunc("/admin/getActors", adminAuthenticate(getActorsHandler))
	http.HandleFunc("/admin/findFilmsByActor", adminAuthenticate(findFilmByActorHandler))
	http.HandleFunc("/admin/findFilmsBySubstring", adminAuthenticate(findFilmByNameHandler))

	// нужно для добавления актеров в актерский состав фильма, если не было фильма, то сначала добавляет
	http.HandleFunc("/admin/addActorsToFilm", adminAuthenticate(addActorsToFilmHandler))

	// user
	http.HandleFunc("/user/getFilms", userAuthenticate(getFilmsHandler))
	http.HandleFunc("/user/getActors", userAuthenticate(getActorsHandler))
	http.HandleFunc("/user/findFilmsBySubstring", userAuthenticate(findFilmByNameHandler))
	http.HandleFunc("/user/findFilmsByActor", userAuthenticate(findFilmByActorHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
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

func checkFilm(film Film) bool {
	return checkFilmName(film.name) && checkFilmDescription(film.description) && checkFilmRating(film.rating)
}

func checkChangedFilm(film changedFilm) bool {
	// по сути каждая скобка это -> (булево следсвтие, если первое верно, то должно быть и второе)
	return (!film.ratingChanged || checkFilmRating(film.newRating)) &&
		(!film.nameChanged || checkFilmName(film.newName)) &&
		(!film.ratingChanged || checkFilmRating(film.newRating)) &&
		checkFilmName(film.prevName)
}

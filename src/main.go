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
	Name string    `json:"name"`
	Sex  bool      `json:"sex"`
	Born time.Time `json:"born"`
}

type Film struct {
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Presentation time.Time `json:"presentation"`
	Rating       int       `json:"rating"`
}

type ChangedActor struct {
	PrevName    string    `json:"prevName"`
	PrevBorn    time.Time `json:"prevBorn"`
	PrevSex     bool      `json:"prevSex"`
	NameChanged bool      `json:"nameChanged"`
	NewName     string    `json:"newName"`
	SexChanged  bool      `json:"sexChanged"`
	NewSex      bool      `json:"newSex"`
	BornChanged bool      `json:"bornChanged"`
	NewBorn     time.Time `json:"newBorn"`
}

type ChangedFilm struct {
	PrevName            string    `json:"prevName"`
	PrevPresentation    time.Time `json:"prevPresentation"`
	NameChanged         bool      `json:"nameChanged"`
	NewName             string    `json:"newName"`
	DescriptionChanged  bool      `json:"descriptionChanged"`
	NewDescription      string    `json:"newDescription"`
	PresentationChanged bool      `json:"presentationChanged"`
	NewPresentation     time.Time `json:"newPresentation"`
	RatingChanged       bool      `json:"ratingChanged"`
	NewRating           int       `json:"newRating"`
}

type SortFilmsParameter struct {
	Sort string `json:"sort"`
}

type SubstringToFind struct {
	Substring string `json:"substring"`
}
type ActorsOfFilm struct {
	Film   Film    `json:"film"`
	Actors []Actor `json:"actors"`
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

	// нужно для добавления актеров в актерский состав фильма, если не было фильма, то сначала добавляет его
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
	return checkFilmName(film.Name) && checkFilmDescription(film.Description) && checkFilmRating(film.Rating)
}

func checkChangedFilm(film ChangedFilm) bool {
	// по сути каждая скобка это -> (булево следсвтие, если первое верно, то должно быть и второе)
	return (!film.RatingChanged || checkFilmRating(film.NewRating)) &&
		(!film.NameChanged || checkFilmName(film.NewName)) &&
		(!film.DescriptionChanged || checkFilmDescription(film.NewDescription)) &&
		checkFilmName(film.PrevName)
}

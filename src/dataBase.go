package main

import (
	"database/sql"
	_ "embed"
	"fmt"
)

type Sex = bool

const (
	Male   Sex = true
	Female Sex = false
)

type LocalStorage interface {
	addFilm(film Film) error
	addActor(actor Actor) error

	deleteFilm(film Film) error
	deleteActor(actor Actor) error

	changeFilm() error
	changeActor() error

	getFilms() ([]Film, error)
	getActors() (map[Actor][]Film, error)

	findFilmsByName() ([]Film, error)
	findFilmsByActor() ([]Film, error)

	addActorsToFilm(actors []Actor, film Film) error
}

type DataBase struct {
	DB    *sql.DB
	Names TablesNames
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
}

type TablesNames struct {
	Films       string
	Actors      string
	FilmsActors string
}

//go:embed migrations/init.sql
var initScript string

func NewDatabase(cfg DBConfig, names TablesNames) (*DataBase, error) {
	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DB)
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(initScript); err != nil {
		return nil, err
	}
	return &DataBase{db, names}, nil
}

func (d *DataBase) addFilm(film Film) error {
	query := fmt.Sprintf("INSERT INTO %s (film_name, description, presentation, rating) VALUES ($1, $2, $3, $4)", d.Names.Films)
	_, err := d.DB.Exec(query, film.name, film.description, film.presentation, film.rating)
	return err
}

func (d *DataBase) addActor(actor Actor) error {
	query := fmt.Sprintf("INSERT INTO %s (actor_name, sex, born) VALUES ($1, $2, $3)", d.Names.Actors)
	_, err := d.DB.Exec(query, actor.name, actor.sex, actor.name)
	return err
}

func (d *DataBase) deleteFilm(film Film) error {
	query := fmt.Sprintf("SELECT film_id FROM %s WHERE film_name = $1 AND presentation = $2", d.Names.Films)
	row := d.DB.QueryRow(query, film.name, film.presentation)

	// получаем id фильма
	var filmId int64
	if err := row.Scan(&filmId); err != nil {
		return err
	}

	// удаляем из таблицы фильм
	query = fmt.Sprintf("DELETE FROM %s WHERE film_name = $1 AND presentation = $2", d.Names.Films)
	if _, err := d.DB.Exec(query, film.name, film.description); err != nil {
		return err
	}

	// удаялем из смежной таблицы актеров и фильмов
	query = fmt.Sprintf("DELETE FROM %s WHERE film_id = $1", d.Names.FilmsActors)
	_, err := d.DB.Exec(query, filmId)

	return err
}

func (d *DataBase) deleteActor(actor Actor) error {
	query := fmt.Sprintf("SELECT actor_id FROM %s WHERE actor_name = $1 AND sex = $2 AND born = $3", d.Names.Actors)
	row := d.DB.QueryRow(query, actor.name, actor.sex, actor.born)

	// получаем id актера
	var actorId int64
	if err := row.Scan(&actorId); err != nil {
		return err
	}

	// удаляем из таблицы актера
	query = fmt.Sprintf("DELETE FROM %s WHERE actor_name = $1 AND sex = $2 AND born = $3", d.Names.Actors)
	if _, err := d.DB.Exec(query, actor.name, actor.sex, actor.born); err != nil {
		return err
	}

	// удаялем из смежной таблицы актеров и фильмов
	query = fmt.Sprintf("DELETE FROM %s WHERE actor_id = $1", d.Names.FilmsActors)
	_, err := d.DB.Exec(query, actorId)

	return err
}

func (d *DataBase) changeFilm() error {
	//TODO implement me
	panic("implement me")
}

func (d *DataBase) changeActor() error {
	//TODO implement me
	panic("implement me")
}

func (d *DataBase) getFilms() ([]Film, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataBase) getActors() (map[Actor][]Film, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataBase) findFilmsByName() ([]Film, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataBase) findFilmsByActor() ([]Film, error) {
	//TODO implement me
	panic("implement me")
}
func (d *DataBase) addActorsToFilm(actors []Actor, film Film) error {
	//TODO implement me
	panic("implement me")
}

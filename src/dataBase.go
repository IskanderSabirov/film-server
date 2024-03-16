package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"sort"
	_ "sort"
	"time"
)

type LocalStorage interface {
	addFilm(film Film) error
	addActor(actor Actor) error

	deleteFilm(film Film) error
	deleteActor(actor Actor) error

	changeFilm() error
	changeActor() error

	getFilms(sort string) ([]Film, error)
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
	query := fmt.Sprintf(
		`INSERT INTO %s (name, description, presentation, rating) 
				VALUES ($1, $2, $3, $4) 
				ON CONFLICT (name, presentation) DO NOTHING`,
		d.Names.Films)
	_, err := d.DB.Exec(query, film.name, film.description, film.presentation, film.rating)
	return err
}

func (d *DataBase) addActor(actor Actor) error {
	query := fmt.Sprintf(
		`INSERT INTO %s (name, sex, born) 
				VALUES ($1, $2, $3)
				ON CONFLICT (name, sex, born) DO NOTHING`,
		d.Names.Actors,
	)
	_, err := d.DB.Exec(query, actor.name, actor.sex, actor.born)
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

func (d *DataBase) getFilms(sortParam string) ([]Film, error) {
	query := fmt.Sprintf(`SELECT * FROM %s`, d.Names.Films)
	rows, err := d.DB.Query(query)
	if err != nil {
		return nil, err
	}
	var films []Film
	var name, description string
	var rating int
	var presentation time.Time

	for rows.Next() {
		if err := rows.Scan(&name, &description, &presentation, rating); err != nil {
			return nil, err
		}
		film := Film{name, description, presentation, rating}
		films = append(films, film)
	}

	switch sortParam {
	case sortByName:
		sort.Slice(films, func(i, j int) bool {
			return films[i].name < films[j].name
		})
		break
	case sortByPresentation:
		sort.Slice(films, func(i, j int) bool {
			return films[i].presentation.Before(films[j].presentation)
		})
		break
	default:
		sort.Slice(films, func(i, j int) bool {
			return films[i].rating < films[j].rating
		})
	}

	return films, nil
}

func (d *DataBase) getActors() (map[Actor][]Film, error) {
	query := `SELECT actors.name, actors.sex, actors.born, films.name, films.description, films.presentation, films.rating
 			  FROM actors
          	  LEFT JOIN films_actors ON actors.id = films_actors.actor
          	  LEFT JOIN films ON films_actors.film = films.id;`

	rows, err := d.DB.Query(query)
	if err != nil {
		return nil, err
	}

	actorsMap := make(map[Actor][]Film)

	var actorName, filmName, filmDescription string
	var actorSex bool
	var filmRating int
	var actorBorn, filmPresentation time.Time

	for rows.Next() {
		if err := rows.Scan(&actorName, &actorSex, &actorBorn, &filmName, &filmDescription, &filmPresentation, &filmRating); err != nil {
			return nil, err
		}
		actor := Actor{actorName, actorSex, actorBorn}
		film := Film{filmName, filmDescription, filmPresentation, filmRating}
		currentFilms, _ := actorsMap[actor]
		currentFilms = append(currentFilms, film)
		actorsMap[actor] = currentFilms
	}

	return actorsMap, nil
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

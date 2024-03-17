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
	addFilm(film Film) (int64, error)
	addActor(actor Actor) (int64, error)

	deleteFilm(film Film) error
	deleteActor(actor Actor) error

	changeFilm(updateFilm changedFilm) error
	changeActor(updateActor changedActor) error

	getFilms(sort string) ([]Film, error)
	getActors() (map[Actor][]Film, error)

	findFilmsByName(nameFragment string) ([]Film, error)
	findFilmsByActor(nameFragment string) ([]Film, error)

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

func (d *DataBase) addFilm(film Film) (int64, error) {
	query := fmt.Sprintf(
		`INSERT INTO %s (name, description, presentation, rating) 
				VALUES ($1, $2, $3, $4) 
				ON CONFLICT (name, presentation) DO NOTHING`,
		d.Names.Films)

	if _, err := d.DB.Exec(query, film.name, film.description, film.presentation, film.rating); err != nil {
		return -1, err
	}

	return d.getFilmID(film)
}

func (d *DataBase) addActor(actor Actor) (int64, error) {
	query := fmt.Sprintf(
		`INSERT INTO %s (name, sex, born)
				VALUES ($1, $2, $3)
				ON CONFLICT (name, sex, born) DO NOTHING`,
		d.Names.Actors,
	)
	if _, err := d.DB.Exec(query, actor.name, actor.sex, actor.born); err != nil {
		return -1, err
	}

	return d.getActorID(actor)
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

func (d *DataBase) changeFilm(film changedFilm) error {
	filmID, err := d.getFilmID(Film{film.prevName, "", film.prevPresentation, 0})
	if err != nil {
		return err
	}

	query := `UPDATE %s
		 	  SET %s = $1
		 	  WHERE EXISTS (SELECT 1 FROM %s WHERE id = %d);`

	if film.nameChanged {
		if _, err := d.DB.Exec(fmt.Sprintf(query, d.Names.Films, "name", d.Names.Films, filmID), film.newName); err != nil {
			return err
		}
	}

	if film.presentationChanged {
		if _, err := d.DB.Exec(fmt.Sprintf(query, d.Names.Films, "presentation", d.Names.Films, filmID), film.newPresentation); err != nil {
			return err
		}
	}

	if film.descriptionChanged {
		if _, err := d.DB.Exec(fmt.Sprintf(query, d.Names.Films, "description", d.Names.Films, filmID), film.newDescription); err != nil {
			return err
		}
	}

	if film.ratingChanged {
		if _, err := d.DB.Exec(fmt.Sprintf(query, d.Names.Films, "rating", d.Names.Films, filmID), film.newRating); err != nil {
			return err
		}
	}

	return nil
}

func (d *DataBase) changeActor(actor changedActor) error {
	actorID, err := d.getActorID(Actor{actor.prevName, actor.prevSex, actor.prevBorn})
	if err != nil {
		return err
	}

	query := `UPDATE %s
		 	  SET %s = $1
		 	  WHERE EXISTS (SELECT 1 FROM %s WHERE id = %d);`

	if actor.bornChanged {
		if _, err := d.DB.Exec(fmt.Sprintf(query, d.Names.Actors, "born", d.Names.Actors, actorID), actor.newBorn); err != nil {
			return err
		}
	}

	if actor.sexChanged {
		if _, err := d.DB.Exec(fmt.Sprintf(query, d.Names.Actors, "sex", d.Names.Actors, actorID), actor.newSex); err != nil {
			return err
		}
	}

	if actor.nameChanged {
		if _, err := d.DB.Exec(fmt.Sprintf(query, d.Names.Actors, "name", d.Names.Actors, actorID), actor.newName); err != nil {
			return err
		}
	}

	return nil
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

	var actorName string
	var filmName, filmDescription sql.NullString
	var actorSex bool
	var filmRating sql.NullInt64
	var actorBorn time.Time
	var filmPresentation sql.NullTime

	for rows.Next() {
		if err := rows.Scan(&actorName, &actorSex, &actorBorn, &filmName, &filmDescription, &filmPresentation, &filmRating); err != nil {
			return nil, err
		}
		actor := Actor{actorName, actorSex, actorBorn}
		if filmName.Valid && filmDescription.Valid && filmRating.Valid && filmPresentation.Valid {
			film := Film{
				filmName.String,
				filmDescription.String,
				filmPresentation.Time,
				int(filmRating.Int64),
			}
			currentFilms, _ := actorsMap[actor]
			currentFilms = append(currentFilms, film)
			actorsMap[actor] = currentFilms
		} else {
			if _, ok := actorsMap[actor]; !ok {
				actorsMap[actor] = []Film{}
			}
		}

	}

	return actorsMap, nil
}

func (d *DataBase) findFilmsByName(nameFragment string) ([]Film, error) {
	query := `SELECT *
			  FROM films 
			  WHERE name LIKE '%' || $1 || '%'`

	rows, err := d.DB.Query(query, nameFragment)
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

	return films, nil

}

func (d *DataBase) findFilmsByActor(nameFragment string) ([]Film, error) {
	query :=
		`SELECT DISTINCT films.name, films.description, films.presentation, films.rating
        FROM actors
               JOIN films_actors ON actors.id = films_actors.actor
               JOIN films ON films_actors.film = films.id
        WHERE actors.name LIKE '%' || $1 || '%'`

	rows, err := d.DB.Query(query, nameFragment)
	if err != nil {
		return nil, err
	}

	var films []Film
	var name, description string
	var presentation time.Time
	var rating int

	for rows.Next() {
		if err := rows.Scan(&name, &description, &presentation, rating); err != nil {
			return nil, err
		}
		film := Film{name, description, presentation, rating}
		films = append(films, film)

	}

	return films, nil
}

func (d *DataBase) addActorsToFilm(actors []Actor, film Film) error {

	filmID, err := d.addFilm(film)
	if err != nil {
		return err
	}
	for _, actor := range actors {
		actorID, err := d.addActor(actor)
		if err != nil {
			return err
		}

		query := fmt.Sprintf(
			`INSERT INTO %s (actor, film)
					VALUES ($1, $2)
					ON CONFLICT (actor, film) DO NOTHING`,
			d.Names.FilmsActors,
		)
		if _, err = d.DB.Exec(query, actorID, filmID); err != nil {
			return err
		}

	}

	return nil
}

func (d *DataBase) getActorID(actor Actor) (int64, error) {
	query := fmt.Sprintf(`SELECT id FROM %s WHERE name=$1, sex=$2, born=$3`, d.Names.Actors)
	row, err := d.DB.Query(query, actor.name, actor.sex, actor.born)

	var id int64
	if err = row.Scan(&id); err != nil {
		return -1, err
	}

	return id, err
}

func (d *DataBase) getFilmID(film Film) (int64, error) {
	query := fmt.Sprintf(`SELECT id FROM %s WHERE name=$1, presentation=$2;`, d.Names.Films)
	row, err := d.DB.Query(query, film.name, film.presentation)

	var id int64
	if err = row.Scan(&id); err != nil {
		return -1, err
	}

	return id, err
}

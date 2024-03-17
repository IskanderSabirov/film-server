package main

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"sort"
	_ "sort"
	"time"
)

//TODO добавить везде логи

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

	getUserPassword(login string) (string, error)
	getAdminPassword(login string) (string, error)
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
	Users       string
	Admins      string
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

	if _, err := d.DB.Exec(query, film.Name, film.Description, film.Presentation, film.Rating); err != nil {
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
	if _, err := d.DB.Exec(query, actor.Name, actor.Sex, actor.Born); err != nil {
		return -1, err
	}

	return d.getActorID(actor)
}

func (d *DataBase) deleteFilm(film Film) error {

	filmId, err := d.getFilmID(film)
	if err != nil {
		return err
	}

	// удаляем из таблицы фильм
	query := fmt.Sprintf("DELETE FROM %s WHERE film_name = $1 AND presentation = $2", d.Names.Films)
	if _, err := d.DB.Exec(query, film.Name, film.Description); err != nil {
		return err
	}

	// удаялем из смежной таблицы актеров и фильмов
	query = fmt.Sprintf("DELETE FROM %s WHERE film_id = $1", d.Names.FilmsActors)
	_, err = d.DB.Exec(query, filmId)

	return err
}

func (d *DataBase) deleteActor(actor Actor) error {
	actorId, err := d.getActorID(actor)
	if err != nil {
		return err
	}

	// удаляем из таблицы актера
	query := fmt.Sprintf("DELETE FROM %s WHERE actor_name = $1 AND sex = $2 AND born = $3", d.Names.Actors)
	if _, err := d.DB.Exec(query, actor.Name, actor.Sex, actor.Born); err != nil {
		return err
	}

	// удаялем из смежной таблицы актеров и фильмов
	query = fmt.Sprintf("DELETE FROM %s WHERE actor_id = $1", d.Names.FilmsActors)
	_, err = d.DB.Exec(query, actorId)

	return err
}

func (d *DataBase) changeFilm(film changedFilm) error {
	filmID, err := d.getFilmID(Film{film.PrevName, "", film.PrevPresentation, 0})
	if err != nil {
		return err
	}

	query := `UPDATE %s
		 	  SET %s = $1
		 	  WHERE EXISTS (SELECT 1 FROM %s WHERE id = %d);`

	if film.NameChanged {
		if _, err := d.DB.Exec(fmt.Sprintf(query, d.Names.Films, "name", d.Names.Films, filmID), film.NewName); err != nil {
			return err
		}
	}

	if film.PresentationChanged {
		if _, err := d.DB.Exec(fmt.Sprintf(query, d.Names.Films, "presentation", d.Names.Films, filmID), film.NewPresentation); err != nil {
			return err
		}
	}

	if film.DescriptionChanged {
		if _, err := d.DB.Exec(fmt.Sprintf(query, d.Names.Films, "description", d.Names.Films, filmID), film.NewDescription); err != nil {
			return err
		}
	}

	if film.RatingChanged {
		if _, err := d.DB.Exec(fmt.Sprintf(query, d.Names.Films, "rating", d.Names.Films, filmID), film.NewRating); err != nil {
			return err
		}
	}

	return nil
}

func (d *DataBase) changeActor(actor changedActor) error {
	actorID, err := d.getActorID(Actor{actor.PrevName, actor.PrevSex, actor.PrevBorn})
	if err != nil {
		return err
	}

	query := `UPDATE %s
		 	  SET %s = $1
		 	  WHERE EXISTS (SELECT 1 FROM %s WHERE id = %d);`

	if actor.BornChanged {
		if _, err := d.DB.Exec(fmt.Sprintf(query, d.Names.Actors, "born", d.Names.Actors, actorID), actor.NewBorn); err != nil {
			return err
		}
	}

	if actor.SexChanged {
		if _, err := d.DB.Exec(fmt.Sprintf(query, d.Names.Actors, "sex", d.Names.Actors, actorID), actor.NewSex); err != nil {
			return err
		}
	}

	if actor.NameChanged {
		if _, err := d.DB.Exec(fmt.Sprintf(query, d.Names.Actors, "name", d.Names.Actors, actorID), actor.NewName); err != nil {
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
			return films[i].Name < films[j].Name
		})
		break
	case sortByPresentation:
		sort.Slice(films, func(i, j int) bool {
			return films[i].Presentation.Before(films[j].Presentation)
		})
		break
	default:
		sort.Slice(films, func(i, j int) bool {
			return films[i].Rating < films[j].Rating
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
	row, err := d.DB.Query(query, actor.Name, actor.Sex, actor.Born)

	var id int64

	if row.Next() {
		if err := row.Scan(&id); err != nil {
			return -1, err
		}
	} else {
		return -1, errors.New("no rows found")
	}

	return id, err
}

func (d *DataBase) getFilmID(film Film) (int64, error) {
	query := fmt.Sprintf(`SELECT id FROM %s WHERE name=$1, presentation=$2;`, d.Names.Films)
	row, err := d.DB.Query(query, film.Name, film.Presentation)

	var id int64

	if row.Next() {
		if err := row.Scan(&id); err != nil {
			return -1, err
		}
	} else {
		return -1, errors.New("no rows found")
	}

	return id, err
}

func (d *DataBase) getUserPassword(login string) (string, error) {
	query := fmt.Sprintf(`SELECT password FROM %s WHERE login = $1;`, d.Names.Users)
	row, err := d.DB.Query(query, login)
	if err != nil {
		return "", nil
	}

	var password string

	if row.Next() {
		if err := row.Scan(&password); err != nil {
			return "", err
		}
	} else {
		return "", errors.New("no rows found")
	}

	return password, nil

}

func (d *DataBase) getAdminPassword(login string) (string, error) {
	query := fmt.Sprintf(`SELECT password FROM %s WHERE login = $1;`, d.Names.Admins)
	row, err := d.DB.Query(query, login)
	if err != nil {
		return "", nil
	}

	var password string

	if row.Next() {
		if err := row.Scan(&password); err != nil {
			return "", err
		}
	} else {
		return "", errors.New("no rows found")
	}

	return password, nil
}

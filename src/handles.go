package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	writingAnswerError string = "Writing answer error, all data added to system"
)

func userAuthenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		log.Printf("Got request with user rights login: '%s' and password: '%s'\n", user, pass)

		password, err := dataBase.getUserPassword(user)
		if err != nil {
			log.Printf("Error in getting password of user\n")

			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Server error in getting user`s password", http.StatusInternalServerError)
		}

		if pass == password {
			log.Printf("Successfully authorized with login:'%s', password:'%s'\n", user, password)
			next(w, r)
		} else {
			log.Printf("Error in password needed:'%s', got:'%s'\n", password, pass)

			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	}
}

func adminAuthenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin, pass, _ := r.BasicAuth()

		log.Printf("Got request with administrator rights login: '%s' and password: '%s'\n", admin, pass)

		password, err := dataBase.getAdminPassword(admin)
		if err != nil {
			log.Printf("Error in getting password of admin\n")

			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Server error in getting admin`s password", http.StatusInternalServerError)
		}

		if pass == password {
			log.Printf("Successfully authorized with login:'%s', password:'%s'\n", admin, password)
			next(w, r)
		} else {
			log.Printf("Error in password needed:'%s', got:'%s'\n", password, pass)
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	}
}

func addActorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var actor Actor

	err := json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Correct decode of json to Actor struct %v\n", actor)

	if _, err := dataBase.addActor(actor); err != nil {
		http.Error(w, "Error in adding actor to data base", http.StatusInternalServerError)
		return
	}

	log.Printf("Correct adding Actor to data base %v\n", actor)

	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprint(w, "Actor successfully added to system\n "); err != nil {
		http.Error(w, writingAnswerError, http.StatusInternalServerError)
		return
	}

}

func changeActorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var changeActor changedActor

	err := json.NewDecoder(r.Body).Decode(&changeActor)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Correct decode of json to changes of actor struct %v\n", changeActor)

	if err := dataBase.changeActor(changeActor); err != nil {
		http.Error(w, "Error in changing information of actor in data base", http.StatusInternalServerError)
		return
	}

	log.Printf("Correct changin information of actor in data base %v\n", changeActor)

	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprint(w, "Actor information successfully changed\n "); err != nil {
		http.Error(w, writingAnswerError, http.StatusInternalServerError)
		return
	}
}

func deleteActorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var actor Actor

	err := json.NewDecoder(r.Body).Decode(&actor)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Correct decode of json to Actor struct %v\n", actor)

	if err := dataBase.deleteActor(actor); err != nil {
		http.Error(w, "Error in deleting actor from data base", http.StatusInternalServerError)
		return
	}

	log.Printf("Correct delete Actor from data base %v\n", actor)

	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprint(w, "Actor information successfully deleted\n "); err != nil {
		http.Error(w, writingAnswerError, http.StatusInternalServerError)
		return
	}
}

func addFilmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var film Film

	err := json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Correct decode of json to Film struct %v\n", film)

	if !checkFilm(film) {
		log.Printf("Incorrect film parameters %v\n", film)
		http.Error(w, "Incorrect film parameters", http.StatusBadRequest)
		return
	}

	if _, err := dataBase.addFilm(film); err != nil {
		http.Error(w, "Error in adding film to data base", http.StatusInternalServerError)
		return
	}

	log.Printf("Correct adding Film to data base %v\n", film)

	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprint(w, "Film successfully added to system\n "); err != nil {
		http.Error(w, writingAnswerError, http.StatusInternalServerError)
		return
	}

}

func changeFilmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var changeFilm changedFilm

	err := json.NewDecoder(r.Body).Decode(&changeFilm)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Correct decode of json to changes of film struct %v\n", changeFilm)

	if !checkChangedFilm(changeFilm) {
		log.Printf("Incorrect change film parameters %v\n", changeFilm)
		http.Error(w, "Incorrect change film parameters ", http.StatusBadRequest)
		return
	}

	if err := dataBase.changeFilm(changeFilm); err != nil {
		http.Error(w, "Error in changing information of film in data base", http.StatusInternalServerError)
		return
	}

	log.Printf("Correct changing information of film in data base %v\n", changeFilm)

	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprint(w, "Film information successfully changed\n "); err != nil {
		http.Error(w, writingAnswerError, http.StatusInternalServerError)
		return
	}
}

func deleteFilmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var film Film

	err := json.NewDecoder(r.Body).Decode(&film)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Correct decode of json to Film struct %v\n", film)

	if err := dataBase.deleteFilm(film); err != nil {
		http.Error(w, "Error in deleting film from data base", http.StatusInternalServerError)
		return
	}

	log.Printf("Correct delete Film from data base %v\n", film)

	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprint(w, "Film information successfully deleted\n "); err != nil {
		http.Error(w, writingAnswerError, http.StatusInternalServerError)
		return
	}
}

func getFilmsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var sortStyle sortFilms

	err := json.NewDecoder(r.Body).Decode(&sortStyle)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Correct decoding of sort method: %s\n", sortStyle.sort)

	films, err := dataBase.getFilms(sortStyle.sort)
	if err != nil {
		http.Error(w, "Error in getting films list from data base", http.StatusInternalServerError)
		return
	}

	log.Printf("Correct receiving films list from data base\n")

	jsonData, err := json.Marshal(films)
	if err != nil {
		http.Error(w, "Server error in encoding films to json", http.StatusInternalServerError)
		return
	}

	log.Printf("Correct encoding films list to json\n")

	w.Header().Set("Content-Type", "application/json")

	if _, err = w.Write(jsonData); err != nil {
		http.Error(w, writingAnswerError, http.StatusInternalServerError)
		return
	}

}

func findFilmByNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var substring findSubstring

	err := json.NewDecoder(r.Body).Decode(&substring)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Correct decoding of substring film`s name: %s\n", substring.substring)

	films, err := dataBase.findFilmsByName(substring.substring)
	if err != nil {
		http.Error(w, "Error in getting films list from data base by name`s substring", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(films)
	if err != nil {
		http.Error(w, "Server error in encoding films to json", http.StatusInternalServerError)
		return
	}

	log.Printf("Correct encoding films list to json\n")

	w.Header().Set("Content-Type", "application/json")

	if _, err = w.Write(jsonData); err != nil {
		http.Error(w, writingAnswerError, http.StatusInternalServerError)
		return
	}
}

func findFilmByActorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var substring findSubstring

	err := json.NewDecoder(r.Body).Decode(&substring)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Correct decoding of actor`s name substring: %s\n", substring.substring)

	films, err := dataBase.findFilmsByActor(substring.substring)
	if err != nil {
		http.Error(w, "Error in getting films list from data base by actor`s name substring", http.StatusInternalServerError)
		return
	}

	log.Printf("Correct getting films by actor`s name substring\n")

	jsonData, err := json.Marshal(films)
	if err != nil {
		http.Error(w, "Server error in encoding films to json", http.StatusInternalServerError)
		return
	}

	log.Printf("Correct encoding films list to json\n")

	w.Header().Set("Content-Type", "application/json")

	if _, err = w.Write(jsonData); err != nil {
		http.Error(w, writingAnswerError, http.StatusInternalServerError)
		return
	}
}

func getActorsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	actors, err := dataBase.getActors()
	if err != nil {
		http.Error(w, "Error in getting actors list from data base", http.StatusInternalServerError)
		return
	}

	log.Printf("Correct receiving actors list from data base\n")

	jsonData, err := json.Marshal(actors)
	if err != nil {
		http.Error(w, "Server error in encoding actors list to json", http.StatusInternalServerError)
		return
	}

	log.Printf("Correct encoding actors list to json\n")

	w.Header().Set("Content-Type", "application/json")

	if _, err = w.Write(jsonData); err != nil {
		http.Error(w, writingAnswerError, http.StatusInternalServerError)
		return
	}
}

func addActorsToFilmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var adding addActorsToFilm

	err := json.NewDecoder(r.Body).Decode(&adding)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Correct decode of json of actors in film struct %v\n", adding)

	if err := dataBase.addActorsToFilm(adding.actors, adding.film); err != nil {
		http.Error(w, "Error in getting actors list from data base", http.StatusInternalServerError)
		return
	}

	log.Printf("Correct added actors to film im data base\n")

	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprint(w, "Actors successfully added to film\n "); err != nil {
		http.Error(w, writingAnswerError, http.StatusInternalServerError)
		return
	}
}

package main

import (
	"testing"
)

func TestCheckFilmName(t *testing.T) {

	var strMaximal [maximumFilmName]byte
	maximal := string(strMaximal[:])

	var strBiggerMaximal [maximumFilmName + 1]byte
	biggerThanMaximal := string(strBiggerMaximal[:])

	testCases := []struct {
		filmName string
		expected bool
	}{
		{"", false},
		{"The Lord of the Rings", true},
		{"A", true},
		{maximal, true},
		{biggerThanMaximal, false},
	}

	for _, tc := range testCases {
		result := checkFilmName(tc.filmName)
		if result != tc.expected {
			t.Errorf("Film name: %s, expected: %t, got: %t", tc.filmName, tc.expected, result)
		}
	}
}

func TestCheckFilmRating(t *testing.T) {
	testCases := []struct {
		rating   int
		expected bool
	}{
		{-1, false},
		{0, true},
		{5, true},
		{11, false},
	}

	for _, tc := range testCases {
		result := checkFilmRating(tc.rating)
		if result != tc.expected {
			t.Errorf("Rating: %d, expected: %t, got: %t", tc.rating, tc.expected, result)
		}
	}
}

func TestCheckFilmDescription(t *testing.T) {

	var strMaximal [maximumFilmDescription]byte
	maximal := string(strMaximal[:])

	var strBiggerMaximal [maximumFilmDescription + 1]byte
	biggerThanMaxima := string(strBiggerMaximal[:])

	testCases := []struct {
		description string
		expected    bool
	}{
		{"", true},
		{"Lorem ipsum", true},
		{biggerThanMaxima, false},
		{maximal, true},
	}

	for _, tc := range testCases {
		result := checkFilmDescription(tc.description)
		if result != tc.expected {
			t.Errorf("Description: %s, expected: %t, got: %t", tc.description, tc.expected, result)
		}
	}
}

func TestCheckFilm(t *testing.T) {
	testCases := []struct {
		film     Film
		expected bool
	}{
		{Film{Name: "The Lord of the Rings", Rating: 9, Description: "Trolls"}, true},
		{Film{Name: "", Rating: 9, Description: "One more troll"}, false},
		{Film{Name: "The Lord of the Rings", Rating: 11, Description: "Trolls become humans : detroit"}, false},
		{Film{Name: "The Lord of the Rings", Rating: 9, Description: ""}, true},
	}

	for _, tc := range testCases {
		result := checkFilm(tc.film)
		if result != tc.expected {
			t.Errorf("Film: %+v, expected: %t, got: %t", tc.film, tc.expected, result)
		}
	}
}

func TestCheckChangedFilm(t *testing.T) {

	var strBiggerMaximalName [maximumFilmName + 1]byte
	biggerThanMaximalName := string(strBiggerMaximalName[:])

	var strBiggerMaximalDescr [maximumFilmDescription + 1]byte
	biggerThanMaximaDescr := string(strBiggerMaximalDescr[:])

	testCases := []struct {
		film     ChangedFilm
		expected bool
	}{
		{ChangedFilm{
			NameChanged:   true,
			NewName:       "The Hobbit",
			RatingChanged: true,
			NewRating:     8,
			PrevName:      "The Lord of the Rings"},
			true},
		//пустое имя
		{ChangedFilm{
			NameChanged:   false,
			NewName:       "",
			RatingChanged: true,
			NewRating:     8,
			PrevName:      "The Lord of the Rings"},
			true},
		//некорректный рейтинг
		{ChangedFilm{
			NameChanged:   true,
			NewName:       "The Hobbit",
			RatingChanged: true,
			NewRating:     11,
			PrevName:      "The Lord of the Rings"},
			false},

		{ChangedFilm{
			NameChanged:   true,
			NewName:       "The Hobbit",
			RatingChanged: true,
			NewRating:     5,
			PrevName:      "The Lord of the Rings"},
			true},
		// пустое новое название
		{ChangedFilm{
			NameChanged:   true,
			NewName:       "",
			RatingChanged: false,
			NewRating:     8,
			PrevName:      "The Lord of the Rings"},
			false},
		// большое название
		{ChangedFilm{
			NameChanged:   true,
			NewName:       biggerThanMaximalName,
			RatingChanged: false,
			NewRating:     8,
			PrevName:      "The Lord of the Rings"},
			false},
		// новое большое описание
		{ChangedFilm{
			NameChanged:        true,
			NewName:            "Hobbit",
			RatingChanged:      false,
			NewRating:          8,
			PrevName:           "The Lord of the Rings",
			DescriptionChanged: true,
			NewDescription:     biggerThanMaximaDescr},
			false},
		//не новое большое описание
		{ChangedFilm{
			NameChanged:        true,
			NewName:            "Hobbit",
			RatingChanged:      false,
			NewRating:          8,
			PrevName:           "The Lord of the Rings",
			DescriptionChanged: false,
			NewDescription:     biggerThanMaximaDescr},
			true},
	}

	for _, tc := range testCases {
		result := checkChangedFilm(tc.film)
		if result != tc.expected {
			t.Errorf("Changed Film: %+v, expected: %t, got: %t", tc.film, tc.expected, result)
		}
	}
}

package main

import "errors"

var longFilmDescription = errors.New("Слишком длинное описание фильма")
var shortFilmDescription = errors.New("Пустое описание фильма")

var longFilmName = errors.New("Слишком длинное название фильма")
var shortFilmName = errors.New("Слишом короткое название фильма фильма")

var incorrectRating = errors.New("Неправильная оценка фильма")

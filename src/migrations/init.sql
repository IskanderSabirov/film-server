CREATE TABLE IF NOT EXISTS films
(
    id           serial PRIMARY KEY,
    name         TEXT,
    description  TEXT,
    presentation DATE,
    rating       INTEGER,

    CONSTRAINT key_films UNIQUE (name, presentation)
);

CREATE TABLE IF NOT EXISTS actors
(
    id   serial PRIMARY KEY,
    name TEXT,
    sex  BOOLEAN,
    born DATE,

    CONSTRAINT key_actors UNIQUE (name, sex, born)
);

CREATE TABLE IF NOT EXISTS films_actors
(
    actor INTEGER,
    film  INTEGER,
    CONSTRAINT key_films_actors UNIQUE (actor, film)
);

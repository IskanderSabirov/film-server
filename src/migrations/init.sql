CREATE TABLE IF NOT EXISTS films(
    film_id serial PRIMARY KEY,
    film_name TEXT,
    description TEXT,
    presentation DATE,
    rating INTEGER
);

CREATE TABLE IF NOT EXISTS actors(
    actor_id serial PRIMARY KEY,
    actor_name TEXT,
    sex BOOLEAN,
    born DATE
);

CREATE TABLE IF NOT EXISTS films_actors(
    actor_id INTEGER,
    film_id INTEGER
);

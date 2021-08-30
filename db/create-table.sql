CREATE TABLE url_db
(
    original_url text        NOT NULL UNIQUE PRIMARY KEY,
    short_url    varchar(10) NOT NULL UNIQUE
);

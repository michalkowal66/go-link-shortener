# go-link-shortener v1.0

## About

go-link-shortener is a web application built using Go, and its libraries: Gin and GORM.

The application is a simple link shortener. It can be used either by accessing the url via a web browser or by consuming the exposed API endpoints.

## Building and running

### Requirements

The application reads environmental variables, make sure they're set before running the app. The data it uses is stored inside a MySQL database, which must be set up as explained later.

#### Environmental variables

`DBUSER` - the name of the database user, used in the connection string

`DBPASS` - the password of the database user, used in the connection string

`DOMAIN` - the name of the domain (set it to your domain/subdomain if exposing the app to WWW or put local serving address - e.g. `localhost:8080`).

`SERVING_AT` - the address at which the application will be served (e.g. `localhost:8080`)

#### Database

By default, the application will try to connect to a database named `go_link_shortener` at `localhost:3306` (default port for MariaDB and MySQL - both can be changed by editing the `main.go` file).

The database requires only one table - `links`.
It can be created using the following SQL command:

``` CREATE TABLE `links` (
`id` int(10) unsigned NOT NULL AUTO_INCREMENT,
`long_url` varchar(1024) NOT NULL,
`url_code` varchar(8) NOT NULL,
`created_on` datetime NOT NULL DEFAULT current_timestamp(),
`last_accessed` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
`times_accessed` int(10) unsigned NOT NULL DEFAULT 1,
PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=26 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci```

### Build

To build the application use the following command
`go build .` or `go build ./main.go` while being inside the main directory of the project.

### Run

To start the application simply run the built executable.

## API

The application exposes endpoints allowing to shorten and and access a shortened URL.

### Shorten URL

A URL can be shortened using the `/shorten` endpoint.
The URL to be shortened must be provided in a sanitized form (to escape special characters) through a `longURL` parameter.

Example: `/shorten?longURL=<URL>`

The response in case of a successful operation will look as follows:

```{ "shortURL": "https://shorter.url/ba3adb44" }```

The trailing part of the short URL is the long URL's ID. The ID can be then used to retrieve the long URL back from the database. 

### Access link

A long URL can be retrieved from the database using the index endpoint (`/`). One simply needs to provide an ID of the long URL to the endpoint.

Example: `/ba3adb44`

The response in case of a successful operation will look as follows:

```{ "longURL": "https://long.url" }```

### Errors

In case of an error, the application will return a message in `json` format with a similar content:

```{ "error": "error message" }```

# Go-HackerNews

*Go-HackerNews* is a [Hacker News](https://news.ycombinator.com) clone written in Go.

## Installation

Inside the project directory install it as a standard Go project with

```shell
$ go install -o go-hackernews ./cmd/web
```

## Usage

Go-HackerNews uses [MySQL](https://www.mysql.com/) as database, which has to be installed and setup before starting Go-HackerNews.

The data source name (DSN) must be provided by the cofiguration file `config.yml`. Inside this file also host and port number are defined.

Start the web server with

```shell
$ go-hackernews
```

## Docker Compose

A `docker-compose.yml` file is provided that starts up both the web server and a MariaDB instance.

You can build it with

```shell
$ docker-compose build
```

and start it with

```shell
$ docker-compose up -d
```

The web server will listen on port 5000, which can be changed in the `docker-compose.yml` file.

## Configuration

You can among other configure the site name and theme colour by adjusting the values found inside `config.yml`.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)
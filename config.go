package main

type config struct {
	Port      int    `env:"TODO_PORT" envDefault:"7540"`
	WebFolder string `envDefault:"./web"`
	DBPath    string `env:"TODO_DBFILE"`
}

package config

type Config struct {
	Port      int    `env:"TODO_PORT" envDefault:"7540"`
	WebFolder string `envDefault:"./web"`
	DBPath    string `env:"TODO_DBFILE"`
	Limit     int    `env:"LIMIT" envDefault:"50"`
	Password  string `env:"TODO_PASSWORD"`
}

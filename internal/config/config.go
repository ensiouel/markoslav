package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
)

type Config struct {
	Postgres Postgres
	Bot      Bot
}

type Postgres struct {
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
	Port     string `env:"POSTGRES_PORT" env-required:"true"`
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DB       string `env:"POSTGRES_DB" env-required:"true"`
}

type Bot struct {
	Debug     bool    `env:"BOT_DEBUG"`
	Token     string  `env:"BOT_TOKEN" env-required:"true"`
	AdminList []int64 `env:"BOT_ADMIN_LIST" env-required:"true"`
}

func New() Config {
	var conf Config
	err := cleanenv.ReadEnv(&conf)
	if err != nil {
		log.Fatal(err)
	}

	return conf
}

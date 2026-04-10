package core_pgx_pool

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string        `envconfig:"HOST" required:"True"`
	Port     string        `envconfig:"PORT" default:"5432"`
	User     string        `envconfig:"USER" required:"True"`
	Pass     string        `envconfig:"PASSWORD" required:"True"`
	Database string        `envconfig:"DB" required:"True"`
	Timeout  time.Duration `envconfig:"TIMEOUT" required:"True"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("POSTGRES", &config); err != nil {
		return Config{}, fmt.Errorf("process env config %w: ", err)
	}

	return config, nil
}

func NewMustConfig() Config {
	cfg, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get Postgres connection pool config: %w", err)
		panic(err)
	}

	return cfg
}

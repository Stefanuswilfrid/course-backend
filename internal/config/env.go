package config

import (
	"log"
	"os"
	"time"
)

type environmentVariables struct {
	ENV         string
	FrontendUrl string
	ApiPort     string

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDbName   string

	JwtAccessSecret    []byte
	JwtAccessDuration  time.Duration
	JwtRefreshSecret   []byte
	JwtRefreshDuration time.Duration
}

var Env *environmentVariables

func LoadEnv() {
	env := &environmentVariables{}
	var err error

	env.ENV = os.Getenv("ENV")
	if env.ENV == "" {
		log.Fatal("ENV is not set")
	}

	env.PostgresHost = os.Getenv("POSTGRES_HOST")
	env.PostgresPort = os.Getenv("POSTGRES_PORT")
	env.PostgresUser = os.Getenv("POSTGRES_USER")
	env.PostgresPassword = os.Getenv("POSTGRES_PASSWORD")
	env.PostgresDbName = os.Getenv("POSTGRES_DB")

	env.JwtAccessSecret = []byte(os.Getenv("JWT_ACCESS_SECRET"))
	env.JwtAccessDuration, err = time.ParseDuration(os.Getenv("JWT_ACCESS_DURATION"))
	if err != nil && env.ENV != "test" {
		log.Fatal("Fail to parse JWT_ACCESS_DURATION")
	}

	env.JwtRefreshSecret = []byte(os.Getenv("JWT_REFRESH_SECRET"))
	env.JwtRefreshDuration, err = time.ParseDuration(os.Getenv("JWT_REFRESH_DURATION"))
	if err != nil && env.ENV != "test" {
		log.Fatal("Fail to parse JWT_REFRESH_DURATION")
	}
	Env = env

}

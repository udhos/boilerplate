package main

import (
	"github.com/udhos/boilerplate/envconfig"
)

type appConfig struct {
	databaseURI  string
	bogus        string
	databaseCode int
	databaseTidy bool
}

func newConfig(env *envconfig.Env) appConfig {
	return appConfig{
		databaseURI:  env.String("DB_URI", "http://test-db"),
		bogus:        env.String("DB_URI", "http://test-db"), // test cache
		databaseCode: env.Int("DB_CODE", 42),
		databaseTidy: env.Bool("DB_TIDY", false),
	}
}

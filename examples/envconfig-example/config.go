package main

import (
	"log"

	"github.com/udhos/boilerplate/awsconfig"
	"github.com/udhos/boilerplate/envconfig"
)

type appConfig struct {
	databaseURI  string
	bogus        string
	databaseCode int
	databaseTidy bool
}

func newConfig() appConfig {

	awsConfOptions := awsconfig.Options{}

	awsConf, errAwsConf := awsconfig.AwsConfig(awsConfOptions)
	if errAwsConf != nil {
		log.Printf("aws config error: %v", errAwsConf)
	}

	envOptions := envconfig.Options{
		QuerySecretsManager: true,
		QueryParameterStore: true,
		AwsConfig:           awsConf.AwsConfig,
	}

	env := envconfig.New(envOptions)

	return appConfig{
		databaseURI:  env.String("DB_URI", "http://test-db"),
		bogus:        env.String("DB_URI", "http://test-db"), // test cache
		databaseCode: env.Int("DB_CODE", 42),
		databaseTidy: env.Bool("DB_TIDY", false),
	}
}

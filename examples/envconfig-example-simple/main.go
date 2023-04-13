// Package main implements an example for env package.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/udhos/boilerplate/boilerplate"
	"github.com/udhos/boilerplate/envconfig"
	"github.com/udhos/boilerplate/secret"
)

// You should really put all your config in a single location.
type appConfig struct {
	databaseURI  string
	databaseCode int
	databaseTidy bool
}

// Load config.
func newConfig(env *envconfig.Env) appConfig {
	return appConfig{
		databaseURI:  env.String("DB_URI", "http://test-db"),
		databaseCode: env.Int("DB_CODE", 42),
		databaseTidy: env.Bool("DB_TIDY", false),
	}
}

func main() {
	me := filepath.Base(os.Args[0])
	log.Println(boilerplate.LongVersion(me))

	roleArn := os.Getenv("ROLE_ARN")

	log.Printf("ROLE_ARN='%s'", roleArn)

	secretOptions := secret.Options{
		RoleSessionName: me,
		RoleArn:         roleArn,
	}
	secret := secret.New(secretOptions)
	envOptions := envconfig.Options{
		Secret: secret,
	}
	env := envconfig.New(envOptions)

	fmt.Printf("\n")
	fmt.Printf("try setting up values like these before running this app.\n")
	fmt.Printf("\n")
	fmt.Printf("(of course you should store the desired value in the correspondig aws service beforehand.)\n")
	fmt.Printf("\n")
	fmt.Printf("export DB_URI=aws-parameterstore:sa-east-1:/microservice9/mongodb:uri\n")
	fmt.Printf("export DB_URI=aws-parameterstore:us-east-1:/microservice9/mongodb:uri\n")
	fmt.Printf("export DB_URI=aws-s3:us-east-1:acredito,app7/mongodb.yaml:uri\n")
	fmt.Printf("export DB_URI=aws-dynamodb:us-east-1:parameters,parameter,mongodb,value:uri\n")
	fmt.Printf("export DB_URI=aws-lambda:us-east-1:parameters,parameter,mongodb,body:uri\n")
	fmt.Printf("\n")

	cfg := newConfig(env)

	fmt.Printf("\n")
	fmt.Printf("databaseURI: %s\n", cfg.databaseURI)
}

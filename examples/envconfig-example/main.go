// Package main implements an example for env package.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/udhos/boilerplate/boilerplate"
	"github.com/udhos/boilerplate/envconfig"
)

func main() {
	me := filepath.Base(os.Args[0])
	log.Println(boilerplate.LongVersion(me))

	env := envconfig.NewSimple(me)

	//loadConfig(env, "DB_URI", "aws-secretsmanager:us-east-1:database:uri")
	loadConfig(env, "DB_URI", "aws-parameterstore:sa-east-1:/microservice9/mongodb:uri")
	loadConfig(env, "DB_URI", "aws-parameterstore:us-east-1:/microservice9/mongodb:uri")
	loadConfig(env, "DB_URI", "aws-s3:us-east-1:acredito,app7/mongodb.yaml:uri")
	loadConfig(env, "DB_URI", "aws-dynamodb:us-east-1:parameters,parameter,mongodb,value:uri")
	loadConfig(env, "DB_URI", "aws-lambda:us-east-1:parameters,parameter,mongodb,body:uri")
	//loadConfig(env, "DB_URI", "#http::GET,https,ttt.lambda-url.us-east-1.on.aws,/,eyJwYXJhbWV0ZXIiOiJtb25nb2RiIn0=,Bearer secret:uri")
}

func loadConfig(env *envconfig.Env, envKey, envValue string) {

	fmt.Println()
	fmt.Println("--------------------------------")
	fmt.Printf("'%s' = '%s'\n", envKey, envValue)
	fmt.Println()

	// this really should be setup when calling the application, not here
	os.Setenv(envKey, envValue)

	cfg := newConfig(env)
	fmt.Printf("'%s' = '%s' => %#v\n", envKey, envValue, cfg)
}

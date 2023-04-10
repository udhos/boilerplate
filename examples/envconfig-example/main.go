// Package main implements an example for env package.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/udhos/boilerplate/awsconfig"
	"github.com/udhos/boilerplate/envconfig"
)

func main() {

	awsConfOptions := awsconfig.Options{}
	awsConf, errAwsConf := awsconfig.AwsConfig(awsConfOptions)
	if errAwsConf != nil {
		log.Printf("aws config error: %v", errAwsConf)
	}
	envOptions := envconfig.Options{
		AwsConfig: awsConf.AwsConfig,
	}
	env := envconfig.New(envOptions)

	loadConfig(env, "DB_URI", "aws-secretsmanager:us-east-1:database:uri")
	loadConfig(env, "DB_URI", "aws-parameterstore:us-east-1:/microservice9/mongodb:uri")
	loadConfig(env, "DB_URI", "aws-s3:us-east-1:acredito,app7/mongodb.yaml:uri")
	loadConfig(env, "DB_URI", "aws-dynamodb:us-east-1:parameters,parameter,mongodb,value:uri")
	loadConfig(env, "DB_URI", "aws-lambda:us-east-1:parameters,parameter,mongodb,body:uri")
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

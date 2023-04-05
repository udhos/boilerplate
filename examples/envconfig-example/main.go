// Package main implements an example for env package.
package main

import (
	"fmt"
	"os"
)

func main() {
	loadConfig("DB_URI", "aws-secretsmanager:us-east-1:database:uri")
	loadConfig("DB_URI", "aws-parameterstore:us-east-1:/microservice9/mongodb:uri")
}

func loadConfig(envKey, envValue string) {

	fmt.Println()
	fmt.Println("--------------------------------")
	fmt.Printf("'%s' = '%s'\n", envKey, envValue)
	fmt.Println()

	os.Setenv(envKey, envValue)
	cfg := newConfig()
	fmt.Printf("'%s' = '%s' => %#v\n", envKey, envValue, cfg)
}

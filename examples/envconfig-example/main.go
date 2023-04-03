// Package main implements an example for env package.
package main

import (
	"fmt"
	"os"
)

func main() {

	os.Setenv("DB_URI", "aws-secretsmanager:us-east-1:database:uri")

	cfg := newConfig()

	fmt.Printf("configuration: %#v\n", cfg)
}

// Package main implements a sample program.
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/udhos/boilerplate/boilerplate"
	"github.com/udhos/boilerplate/secret"
)

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

	load(secret, "vault::token,dev-only-token,http,localhost,8200,secret/myapp1/mongodb:uri")
}

func load(s *secret.Secret, name string) {
	log.Printf("=============================")
	log.Printf("##### RESULT1: %s: %s\n", name, s.Retrieve(name))
	log.Printf("##### RESULT2: %s: %s\n", name, s.Retrieve(name))
	log.Printf("=============================")
}

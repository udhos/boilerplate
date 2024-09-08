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

	secretOptions := secret.Options{}
	secret := secret.New(secretOptions)

	log.Print("TOKEN: export VAULT=vault::token,dev-only-token,http,localhost,8200,secret/myapp1/mongodb:uri")
	log.Print("ROLE:  export VAULT=vault::,dev-role-iam,http,localhost,8200,secret/myapp1/mongodb:uri")

	v := os.Getenv("VAULT")
	if v == "" {
		v = "vault::token,dev-only-token,http,localhost,8200,secret/myapp1/mongodb:uri"
	}
	log.Printf("VAULT=%s", v)

	load(secret, v)
}

func load(s *secret.Secret, name string) {
	log.Printf("=============================")
	log.Printf("##### RESULT1: %s: %s\n", name, s.Retrieve(name))
	log.Printf("##### RESULT2: %s: %s\n", name, s.Retrieve(name))
	log.Printf("=============================")
}

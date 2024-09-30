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

	debug := os.Getenv("DEBUG")
	secretOptions := secret.Options{Debug: debug != ""}
	secret := secret.New(secretOptions)

	log.Print("TOKEN: export SECRET='proxy||http,localhost,8080,vault::token,dev-only-token,http,localhost,8200,secret/myapp1/mongodb:uri'")
	log.Print("ROLE:  export SECRET='proxy||http,localhost,8080,vault::,dev-role-iam,http,localhost,8200,secret/myapp1/mongodb:uri'")

	v := os.Getenv("SECRET")
	if v == "" {
		v = "proxy||http,localhost,8080,vault::token,dev-only-token,http,localhost,8200,secret/myapp1/mongodb:uri"
	}
	log.Printf("SECRET=%s", v)

	load(secret, v)
}

func load(s *secret.Secret, name string) {
	log.Printf("=============================")
	log.Printf("##### RESULT1: %s: %s\n", name, s.Retrieve(name))
	log.Printf("##### RESULT2: %s: %s\n", name, s.Retrieve(name))
	log.Printf("=============================")
}

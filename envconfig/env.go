// Package envconfig provides utilities for reading environment variables.
package envconfig

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/udhos/boilerplate/boilerplate"
	"github.com/udhos/boilerplate/secret"
)

// Env holds context for loading config env vars.
type Env struct {
	options Options
}

// Options defines client options.
type Options struct {
	DisableQueryStore bool
	Secret            *secret.Secret
	Printf            boilerplate.FuncPrintf
}

// New creates a client for loading config from env vars.
func New(opt Options) *Env {

	if opt.Printf == nil {
		opt.Printf = log.Printf
	}

	return &Env{options: opt}
}

func (e *Env) getEnv(name string) string {
	value := os.Getenv(name)

	if value == "" {
		return value
	}

	if e.options.DisableQueryStore {
		return value
	}

	return e.options.Secret.Retrieve(value)
}

// String extracts string from env var.
// It returns the provided defaultValue if the env var is empty.
// The string returned is also recorded in logs.
func (e *Env) String(name string, defaultValue string) string {
	str := e.getEnv(name)
	if str != "" {
		e.options.Printf("%s=[%s] using %s=%s default=%s", name, str, name, str, defaultValue)
		return str
	}
	e.options.Printf("%s=[%s] using %s=%s default=%s", name, str, name, defaultValue, defaultValue)
	return defaultValue
}

// Bool extracts boolean value from env var.
// It returns the provided defaultValue if the env var is empty.
// The value returned is also recorded in logs.
func (e *Env) Bool(name string, defaultValue bool) bool {
	str := e.getEnv(name)
	if str != "" {
		value, errConv := strconv.ParseBool(str)
		if errConv == nil {
			e.options.Printf("%s=[%s] using %s=%v default=%v", name, str, name, value, defaultValue)
			return value
		}
		e.options.Printf("bad %s=[%s]: error: %v", name, str, errConv)
	}
	e.options.Printf("%s=[%s] using %s=%v default=%v", name, str, name, defaultValue, defaultValue)
	return defaultValue
}

// Duration extracts time.Duration value from env var.
// It returns the provided defaultValue if the env var is empty.
// The value returned is also recorded in logs.
func (e *Env) Duration(name string, defaultValue time.Duration) time.Duration {
	str := e.getEnv(name)
	if str != "" {
		value, errConv := time.ParseDuration(str)
		if errConv == nil {
			e.options.Printf("%s=[%s] using %s=%v default=%v", name, str, name, value, defaultValue)
			return value
		}
		e.options.Printf("bad %s=[%s]: error: %v", name, str, errConv)
	}
	e.options.Printf("%s=[%s] using %s=%v default=%v", name, str, name, defaultValue, defaultValue)
	return defaultValue
}

// Int extracts int value from env var.
// It returns the provided defaultValue if the env var is empty.
// The value returned is also recorded in logs.
func (e *Env) Int(name string, defaultValue int) int {
	str := e.getEnv(name)
	if str != "" {
		value, errConv := strconv.Atoi(str)
		if errConv == nil {
			e.options.Printf("%s=[%s] using %s=%v default=%v", name, str, name, value, defaultValue)
			return value
		}
		e.options.Printf("bad %s=[%s]: error: %v", name, str, errConv)
	}
	e.options.Printf("%s=[%s] using %s=%v default=%v", name, str, name, defaultValue, defaultValue)
	return defaultValue
}

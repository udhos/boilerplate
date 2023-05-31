// Package envconfig provides utilities for reading environment variables.
package envconfig

import (
	"log"
	"os"
	"strconv"
	"strings"
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

// NewSimple creates a client for loading config from env vars.
func NewSimple(sessionName string) *Env {
	roleArn := os.Getenv("SECRET_ROLE_ARN")

	log.Printf("envconfig.NewSimple: SECRET_ROLE_ARN='%s'", roleArn)

	secretOptions := secret.Options{
		RoleSessionName: sessionName,
		RoleArn:         roleArn,
	}
	secret := secret.New(secretOptions)
	envOptions := Options{
		Secret: secret,
	}
	env := New(envOptions)
	return env
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

// Uint64 extracts uint64 value from env var.
// It returns the provided defaultValue if the env var is empty.
// The value returned is also recorded in logs.
func (e *Env) Uint64(name string, defaultValue uint64) uint64 {
	str := e.getEnv(name)
	if str != "" {
		value, errConv := strconv.ParseUint(name, 10, 64)
		if errConv == nil {
			e.options.Printf("%s=[%s] using %s=%v default=%v", name, str, name, value, defaultValue)
			return value
		}
		e.options.Printf("bad %s=[%s]: error: %v", name, str, errConv)
	}
	e.options.Printf("%s=[%s] using %s=%v default=%v", name, str, name, defaultValue, defaultValue)
	return defaultValue
}

// Float64Slice extracts []float64 from env var.
// It returns the provided defaultValue if the env var is empty.
// The value returned is also recorded in logs.
func (e *Env) Float64Slice(name string, defaultValue []float64) []float64 {
	str := e.getEnv(name)
	if str == "" {
		e.options.Printf("%s=[%s] using %s=%v default=%v", name, str, name, defaultValue, defaultValue)
		return defaultValue
	}

	var value []float64
	items := strings.FieldsFunc(str, func(sep rune) bool { return sep == ',' })
	for i, field := range items {
		field = strings.TrimSpace(field)
		f, errConv := strconv.ParseFloat(field, 64)
		if errConv != nil {
			e.options.Printf("bad %s=[%s] error parsing item %d='%s': %v: using %s=%v default=%v",
				name, str, i, field, errConv, name, value, defaultValue)
			return defaultValue
		}
		value = append(value, f)
	}

	e.options.Printf("%s=[%s] using %s=%v default=%v", name, str, name, value, defaultValue)

	return value
}

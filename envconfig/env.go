// Package envconfig provides utilities for reading environment variables.
package envconfig

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// Options provide optional parameters for client.
type Options struct {
	AwsConfig                  aws.Config
	Printf                     FuncPrintf // defaults to e.options.Printf
	PrefixSecretsManager       string     // defaults to "aws-secretsmanager"
	QuerySecretsManager        bool
	CrashOnSecretsManagerError bool
}

const DefaultSecretsManagerPrefix = "aws-secretsmanager"

// FuncPrintf is a helper type for logging function.
type FuncPrintf func(format string, v ...any)

// Env holds context information for loading confing from env vars.
type Env struct {
	options Options
	cache   map[string]secret
}

// New creates a client for loading config from env vars.
func New(opt Options) *Env {

	if opt.Printf == nil {
		opt.Printf = log.Printf
	}

	if opt.PrefixSecretsManager == "" {
		opt.PrefixSecretsManager = DefaultSecretsManagerPrefix
	}

	return &Env{
		options: opt,
		cache:   map[string]secret{},
	}
}

func (e *Env) getEnv(name string) string {
	value := os.Getenv(name)

	if e.options.QuerySecretsManager {
		if value == "" {
			return ""
		}
		value = e.secretsManagerGet(value)
	}

	return value
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

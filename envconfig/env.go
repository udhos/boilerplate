// Package envconfig provides utilities for reading environment variables.
package envconfig

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// Options provide optional parameters for client.
type Options struct {
	AwsConfig            aws.Config
	Printf               FuncPrintf // defaults to log.Printf
	PrefixSecretsManager string     // defaults to "aws-secretsmanager"
	PrefixParameterStore string     // defaults to "aws-parameterstore"
	PrefixS3             string     // defaults to "aws-s3"
	QuerySecretsManager  bool
	QueryParameterStore  bool
	QueryS3              bool
	CrashOnQueryError    bool
}

// Define default prefixes for Secrets Manager and Parameter Store.
const (
	DefaultSecretsManagerPrefix = "aws-secretsmanager"
	DefaultParameterStorePrefix = "aws-parameterstore"
	DefaultS3Prefix             = "aws-s3"
)

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

	if opt.PrefixParameterStore == "" {
		opt.PrefixParameterStore = DefaultParameterStorePrefix
	}

	if opt.PrefixS3 == "" {
		opt.PrefixS3 = DefaultS3Prefix
	}

	return &Env{
		options: opt,
		cache:   map[string]secret{},
	}
}

func (e *Env) getEnv(name string) string {
	value := os.Getenv(name)

	switch {
	case e.options.QuerySecretsManager && strings.HasPrefix(value, e.options.PrefixSecretsManager):
		value = e.query(querySecret, e.options.PrefixSecretsManager, value)
	case e.options.QueryParameterStore && strings.HasPrefix(value, e.options.PrefixParameterStore):
		value = e.query(queryParameter, e.options.PrefixParameterStore, value)
	case e.options.QueryS3 && strings.HasPrefix(value, e.options.PrefixS3):
		value = e.query(queryS3, e.options.PrefixS3, value)
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

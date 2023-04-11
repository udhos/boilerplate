// Package envconfig loads configuration from env vars.
package envconfig

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/udhos/boilerplate/awsconfig"
	"gopkg.in/yaml.v2"
)

// aws-secretsmanager:region:name:json_field
func (e *Env) query(q queryFunc, prefix, key string) string {
	const me = "query"

	//
	// parse key: aws-secretsmanager:region:name:json_field
	//

	fields := strings.SplitN(key, ":", 4)
	if len(fields) < 3 {
		e.options.Printf("%s: missing fields: %s", me, key)
		return key
	}

	if fields[0] != prefix {
		e.options.Printf("%s: missing prefix='%s': %s", me, prefix, key)
		return key
	}

	region := fields[1]
	secretName := fields[2]
	var jsonField string
	if len(fields) > 3 {
		jsonField = fields[3]
	}

	e.options.Printf("%s: key='%s' json_field=%s",
		me, key, jsonField)

	//
	// retrieve secret
	//

	begin := time.Now()

	secretString, errSecret := e.retrieve(q, region, secretName, jsonField)

	e.options.Printf("%s: query: key='%s': elapsed: %v",
		me, key, time.Since(begin))

	if errSecret != nil {
		e.options.Printf("%s: secret error: key='%s': %v",
			me, key, errSecret)
		if e.options.CrashOnQueryError {
			e.options.Printf("%s: crashing on error: key='%s': %v",
				me, key, errSecret)
			os.Exit(1)
		}
		return key
	}

	if jsonField == "" {
		// return scalar (non-JSON) secret
		e.options.Printf("%s: key='%s' json_field=%s: value=%s",
			me, key, jsonField, secretString)
		return secretString
	}

	//
	// extract field from secret in JSON
	//

	value := map[string]string{}

	errJSON := yaml.Unmarshal([]byte(secretString), &value)
	if errJSON != nil {
		e.options.Printf("%s: json error: key='%s': %v",
			me, key, errJSON)
		if e.options.CrashOnQueryError {
			e.options.Printf("%s: crashing on error: key='%s': %v",
				me, key, errJSON)
			os.Exit(1)
		}
		return secretString
	}

	fieldValue := value[jsonField]

	e.options.Printf("%s: key='%s' json_field=%s: value=%s",
		me, key, jsonField, fieldValue)

	return fieldValue
}

//
// We only cache secrets with JSON fields:
//
//     {"uri":"mongodb://127.0.0.2:27017", "database":"bogus"}
//
// In order to fetch multiple fields from a secret with a single (cached)
// query against AWS Secrets Manager:
//
//     export      MONGO_URL=aws-secretsmanager:us-east-1:mongo:uri
//     export MONGO_DATABASE=aws-secretsmanager:us-east-1:mongo:database
//

type secret struct {
	value   string
	created time.Time
}

func (e *Env) retrieve(q queryFunc, region, secretName, field string) (string, error) {
	const (
		me       = "retrieve"
		cacheTTL = time.Minute
	)

	var cacheKey string
	var secretString string

	if field != "" {
		//
		// check cache, only for JSON values
		//
		cacheKey = region + ":" + secretName
		cached, found := e.cache[cacheKey]
		if found {
			// cache hit
			elapsed := time.Since(cached.created)
			if elapsed < cacheTTL {
				// live entry
				secretString = cached.value
				e.options.Printf("%s: from cache: %s=%s (elapsed=%s TTL=%s)",
					me, cacheKey, secretString, elapsed, cacheTTL)
				return secretString, nil
			}
			// stale entry
			delete(e.cache, cacheKey)
		}
	}

	//
	// field not provided || cache miss || stale cache entry
	//

	//
	// retrieve from secrets manager
	//
	e.awsConfSrc.awsConfigOptions.Region = region

	value, errSecret := q(e.awsConfSrc, secretName)
	if errSecret != nil {
		e.options.Printf("%s: secret error: %v", me, errSecret)
		return value, errSecret
	}
	secretString = value

	//
	// retrieved value from service
	//

	e.options.Printf("%s: from secretsmanager: %s=%s", me, secretName, secretString)

	if field != "" {
		//
		// save to cache
		//
		e.cache[cacheKey] = secret{
			value:   secretString,
			created: time.Now(),
		}
	}

	return secretString, nil
}

type awsConfigSource struct {
	awsConfigOptions awsconfig.Options
}

func (s *awsConfigSource) get() (aws.Config, error) {
	output, err := awsconfig.AwsConfig(s.awsConfigOptions)
	return output.AwsConfig, err
}

type awsConfigSolver interface {
	get() (aws.Config, error)
}

type queryFunc func(getAwsConfig awsConfigSolver, name string) (string, error)

/*
func fieldRegion(s string) string {
	fields := strings.SplitN(s, ":", 2)
	if len(fields) < 2 {
		return ""
	}
	return fields[1]
}
*/

func querySecret(getAwsConfig awsConfigSolver, secretName string) (string, error) {
	const me = "querySecret"

	awsConfig, errAwsConfig := getAwsConfig.get()
	if errAwsConfig != nil {
		return "", errAwsConfig
	}

	sm := secretsmanager.NewFromConfig(awsConfig)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}
	result, errSecret := sm.GetSecretValue(context.TODO(), input)
	if errSecret != nil {
		return "", errSecret
	}
	return *result.SecretString, nil
}

// Package envconfig loads configuration from env vars.
package envconfig

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// secretsmanager:region:name:json_field
func (e *Env) secretsManagerGet(key string) string {
	const me = "secretsManagerGet"

	//
	// parse key: secretsmanager:region:name:json_field
	//

	fields := strings.SplitN(key, ":", 4)
	if len(fields) < 3 {
		e.options.Printf("%s: missing fields: %s", me, key)
		return key
	}

	if fields[0] != e.options.PrefixSecretsManager {
		e.options.Printf("%s: missing prefix: %s", me, key)
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

	secretString, errSecret := e.retrieve(region, secretName, jsonField)
	if errSecret != nil {
		e.options.Printf("%s: secret error: key='%s': %v",
			me, key, errSecret)
		if e.options.CrashOnSecretsManagerError {
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

	errJSON := json.Unmarshal([]byte(secretString), &value)
	if errJSON != nil {
		e.options.Printf("%s: json error: key='%s': %v",
			me, key, errJSON)
		if e.options.CrashOnSecretsManagerError {
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
//     export      MONGO_URL=secretsmanager:us-east-1:mongo:uri
//     export MONGO_DATABASE=secretsmanager:us-east-1:mongo:database
//

type secret struct {
	value   string
	created time.Time
}

func (e *Env) retrieve(region, secretName, field string) (string, error) {
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
			elapsed := time.Since(cached.created)
			if elapsed < cacheTTL {
				secretString = cached.value
				e.options.Printf("%s: from cache: %s=%s (elapsed=%s TTL=%s)",
					me, cacheKey, secretString, elapsed, cacheTTL)
			} else {
				delete(e.cache, cacheKey)
			}
		}
	}

	if secretString == "" {
		//
		// load aws config
		//
		sm := secretsmanager.NewFromConfig(e.options.AwsConfig)

		//
		// retrieve from secrets manager
		//
		input := &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(secretName),
			VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
		}
		result, errSecret := sm.GetSecretValue(context.TODO(), input)
		if errSecret != nil {
			e.options.Printf("%s: secret error: %v", me, errSecret)
			return "", errSecret
		}
		secretString = *result.SecretString

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
	}

	return secretString, nil
}

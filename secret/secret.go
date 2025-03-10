// Package secret retrieves secrets.
package secret

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"gopkg.in/yaml.v3"

	"github.com/udhos/boilerplate/awsconfig"
	"github.com/udhos/boilerplate/boilerplate"
)

// Options provide optional parameters for client.
type Options struct {
	Debug                bool
	Printf               boilerplate.FuncPrintf // defaults to log.Printf
	PrefixSecretsManager string                 // defaults to "aws-secretsmanager"
	PrefixParameterStore string                 // defaults to "aws-parameterstore"
	PrefixS3             string                 // defaults to "aws-s3"
	PrefixDynamoDb       string                 // defaults to "aws-dynamodb"
	PrefixLambda         string                 // defaults to "aws-lambda"
	PrefixHTTP           string                 // defaults to "#http"
	PrefixVault          string                 // defaults to "vault"
	PrefixProxy          string                 // defaults to "proxy"
	CrashOnQueryError    bool                   // require secret
	CacheTTLSeconds      int                    // cache TTL in seconds: -1=noCache 0=useDefault (60)
	AwsConfigSource      AwsConfigSolver
}

// Define default prefixes for Secrets Manager and Parameter Store.
const (
	DefaultSecretsManagerPrefix = "aws-secretsmanager"
	DefaultParameterStorePrefix = "aws-parameterstore"
	DefaultS3Prefix             = "aws-s3"
	DefaultDynamoDbPrefix       = "aws-dynamodb"
	DefaultLambdaPrefix         = "aws-lambda"
	DefaultHTTPPrefix           = "#http"
	DefaultVaultPrefix          = "vault"
	DefaultProxyPrefix          = "proxy"
)

// Secret holds context information for retrieving secrets.
type Secret struct {
	options Options
	cache   map[string]secret
}

// New creates a Secret context for retrieving secrets.
func New(opt Options) *Secret {

	if opt.AwsConfigSource == nil {
		panic("AwsConfigSource is nil")
	}

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

	if opt.PrefixDynamoDb == "" {
		opt.PrefixDynamoDb = DefaultDynamoDbPrefix
	}

	if opt.PrefixLambda == "" {
		opt.PrefixLambda = DefaultLambdaPrefix
	}

	if opt.PrefixHTTP == "" {
		opt.PrefixHTTP = DefaultHTTPPrefix
	}

	if opt.PrefixVault == "" {
		opt.PrefixVault = DefaultVaultPrefix
	}

	if opt.PrefixProxy == "" {
		opt.PrefixProxy = DefaultProxyPrefix
	}

	if opt.CacheTTLSeconds < 0 {
		opt.CacheTTLSeconds = 0 // disable cache
	} else if opt.CacheTTLSeconds == 0 {
		opt.CacheTTLSeconds = 60 // default 60 seconds
	}

	return &Secret{
		options: opt,
		cache:   map[string]secret{},
	}
}

// Retrieve fetches a secret.
// If an error is found, only crashes if CrashOnQueryError is set.
// name: aws-secretsmanager:region:name:json_field
func (s *Secret) Retrieve(name string) string {
	const me = "Secret.Retrieve"

	value, err := s.RetrieveWithError(name)
	if err != nil {
		s.options.Printf("%s: error: name='%s': %v",
			me, name, err)
		if s.options.CrashOnQueryError {
			s.options.Printf("%s: error: crashing on error: name='%s': %v",
				me, name, err)
			os.Exit(1)
		}
		return name
	}

	return value
}

// RetrieveWithError fetches a secret.
// name: aws-secretsmanager:region:name:json_field
func (s *Secret) RetrieveWithError(name string) (string, error) {

	var err error

	switch {
	case strings.HasPrefix(name, s.options.PrefixSecretsManager):
		name, err = s.query(querySecret, s.options.PrefixSecretsManager, name)
	case strings.HasPrefix(name, s.options.PrefixParameterStore):
		name, err = s.query(queryParameter, s.options.PrefixParameterStore, name)
	case strings.HasPrefix(name, s.options.PrefixS3):
		name, err = s.query(queryS3, s.options.PrefixS3, name)
	case strings.HasPrefix(name, s.options.PrefixDynamoDb):
		name, err = s.query(queryDynamoDb, s.options.PrefixDynamoDb, name)
	case strings.HasPrefix(name, s.options.PrefixLambda):
		name, err = s.query(queryLambda, s.options.PrefixLambda, name)
	case strings.HasPrefix(name, s.options.PrefixHTTP):
		name, err = s.query(queryHTTP, s.options.PrefixHTTP, name)
	case strings.HasPrefix(name, s.options.PrefixVault):
		name, err = s.query(queryVault, s.options.PrefixVault, name)
	case strings.HasPrefix(name, s.options.PrefixProxy):
		name, err = s.query(queryProxy, s.options.PrefixProxy, name)
	}

	return name, err
}

// querySimple retrieves a secret.
// If an error is found, only crashes if CrashOnQueryError is set.
// key: aws-secretsmanager:region:name:json_field
func (s *Secret) querySimple(q queryFunc, prefix, key string) string {
	const me = "querySimple"

	value, errQuery := s.query(q, prefix, key)

	if errQuery != nil {
		s.options.Printf("%s: error: key='%s': %v",
			me, key, errQuery)
		if s.options.CrashOnQueryError {
			s.options.Printf("%s: error: crashing on error: key='%s': %v",
				me, key, errQuery)
			os.Exit(1)
		}
		return key
	}

	return value
}

func parseSecretName(prefix, name string) (string, string, string, error) {

	const me = "parseSecretName"

	trimPrefix := strings.TrimPrefix(name, prefix)
	if trimPrefix == name {
		return "", "", "", fmt.Errorf("%s: missing prefix='%s': %s", me, prefix, name)
	}
	if len(trimPrefix) < 1 {
		return "", "", "", fmt.Errorf("%s: secret too short length=%d prefix='%s': %s",
			me, len(trimPrefix), prefix, name)
	}

	separator := trimPrefix[:1]

	fields := strings.SplitN(name, separator, 4)
	if len(fields) < 3 {
		return "", "", "", fmt.Errorf("%s: missing fields: %s", me, name)
	}

	if fields[0] != prefix {
		return "", "", "", fmt.Errorf("%s: missing prefix='%s': %s", me, prefix, name)
	}

	region := fields[1]
	secretName := fields[2]
	var jsonField string
	if len(fields) > 3 {
		jsonField = fields[3]
	}

	return region, secretName, jsonField, nil
}

// query retrieves a secret.
// key: aws-secretsmanager:region:name:json_field
func (s *Secret) query(q queryFunc, prefix, key string) (string, error) {
	const me = "query"

	//
	// parse key: aws-secretsmanager:region:name:json_field
	//

	region, secretName, jsonField, errParse := parseSecretName(prefix, key)
	if errParse != nil {
		s.options.Printf("%s: parse secret error: %v", me, errParse)
		return key, nil
	}

	if s.options.Debug {
		s.options.Printf("%s: key='%s' json_field=%s",
			me, key, jsonField)
	}

	//
	// retrieve secret
	//

	begin := time.Now()

	secretString, errSecret := s.retrieve(q, region, secretName, jsonField)

	if s.options.Debug {
		s.options.Printf("%s: query: key='%s': elapsed: %v",
			me, key, time.Since(begin))
	}

	if errSecret != nil {
		s.options.Printf("%s: secret error: key='%s': %v",
			me, key, errSecret)
		return key, errSecret
	}

	if jsonField == "" {
		// return scalar (non-JSON) secret
		if s.options.Debug {
			s.options.Printf("%s: key='%s' json_field=%s: value=%s",
				me, key, jsonField, secretString)
		}
		return secretString, nil
	}

	//
	// extract field from secret in JSON
	//

	value := map[string]string{}

	errJSON := yaml.Unmarshal([]byte(secretString), &value)
	if errJSON != nil {
		s.options.Printf("%s: json error: key='%s': %v",
			me, key, errJSON)
		return secretString, errJSON
	}

	fieldValue := value[jsonField]

	if s.options.Debug {
		s.options.Printf("%s: key='%s' json_field=%s: value=%s",
			me, key, jsonField, fieldValue)
	}

	return fieldValue, nil
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

func (s *Secret) retrieve(q queryFunc, region, secretName, field string) (string, error) {
	const me = "Secret.retrieve"

	var cacheKey string
	var secretString string

	if field != "" {
		//
		// check cache, only for JSON values
		//
		cacheKey = region + ":" + secretName
		cached, found := s.cache[cacheKey]
		if found {
			// cache hit
			elapsed := time.Since(cached.created)
			ttl := time.Second * time.Duration(s.options.CacheTTLSeconds)
			if elapsed < ttl {
				// live entry
				secretString = cached.value
				if s.options.Debug {
					s.options.Printf("%s: from cache: %s=%s (elapsed=%s TTL=%s)",
						me, cacheKey, secretString, elapsed, ttl)
				}
				return secretString, nil
			}
			// stale entry
			delete(s.cache, cacheKey)
		}
	}

	//
	// field not provided || cache miss || stale cache entry
	//

	//
	// retrieve from secrets manager
	//
	s.options.AwsConfigSource.setRegion(region)

	value, errSecret := q(s.options.Debug, s.options.Printf, s.options.AwsConfigSource, secretName)
	if errSecret != nil {
		s.options.Printf("%s: secret query error: %v", me, errSecret)
		return value, errSecret
	}
	secretString = value

	//
	// retrieved value from service
	//
	if s.options.Debug {
		s.options.Printf("%s: from store: %s=%s", me, secretName, secretString)
	}

	if field != "" {
		//
		// save to cache
		//
		s.cache[cacheKey] = secret{
			value:   secretString,
			created: time.Now(),
		}
	}

	return secretString, nil
}

// AwsConfigSource implements AwsConfigSolver.
type AwsConfigSource struct {
	AwsConfigOptions awsconfig.Options
}

func (s *AwsConfigSource) get() (aws.Config, error) {
	output, err := awsconfig.AwsConfig(s.AwsConfigOptions)
	return output.AwsConfig, err
}

func (s *AwsConfigSource) endpointURL() string {
	return s.AwsConfigOptions.EndpointURL
}

func (s *AwsConfigSource) setRegion(region string) {
	s.AwsConfigOptions.Region = region
}

// AwsConfigSolver provides aws configuration.
type AwsConfigSolver interface {
	get() (aws.Config, error)
	endpointURL() string
	setRegion(region string)
}

type queryFunc func(debug bool, printf boilerplate.FuncPrintf, getAwsConfig AwsConfigSolver, name string) (string, error)

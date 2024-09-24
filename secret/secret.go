// Package secret retrieves secrets.
package secret

import (
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
	RoleArn              string
	RoleSessionName      string
	CrashOnQueryError    bool
	CacheTTL             time.Duration // defaults to 1 minute
	EndpointURL          string
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
)

// Secret holds context information for retrieving secrets.
type Secret struct {
	options    Options
	cache      map[string]secret
	awsConfSrc *awsConfigSource
}

// New creates a Secret context for retrieving secrets.
func New(opt Options) *Secret {

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

	if opt.CacheTTL == 0 {
		opt.CacheTTL = time.Minute
	}

	awsConfOptions := awsconfig.Options{
		Printf:          opt.Printf,
		RoleArn:         opt.RoleArn,
		RoleSessionName: opt.RoleSessionName,
		EndpointURL:     opt.EndpointURL,
	}

	return &Secret{
		options:    opt,
		cache:      map[string]secret{},
		awsConfSrc: &awsConfigSource{awsConfigOptions: awsConfOptions},
	}
}

// Retrieve fetches a secret.
func (s *Secret) Retrieve(name string) string {

	switch {
	case strings.HasPrefix(name, s.options.PrefixSecretsManager):
		name = s.query(querySecret, s.options.PrefixSecretsManager, name)
	case strings.HasPrefix(name, s.options.PrefixParameterStore):
		name = s.query(queryParameter, s.options.PrefixParameterStore, name)
	case strings.HasPrefix(name, s.options.PrefixS3):
		name = s.query(queryS3, s.options.PrefixS3, name)
	case strings.HasPrefix(name, s.options.PrefixDynamoDb):
		name = s.query(queryDynamoDb, s.options.PrefixDynamoDb, name)
	case strings.HasPrefix(name, s.options.PrefixLambda):
		name = s.query(queryLambda, s.options.PrefixLambda, name)
	case strings.HasPrefix(name, s.options.PrefixHTTP):
		name = s.query(queryHTTP, s.options.PrefixHTTP, name)
	case strings.HasPrefix(name, s.options.PrefixVault):
		name = s.query(queryVault, s.options.PrefixVault, name)
	}

	return name
}

// aws-secretsmanager:region:name:json_field
func (s *Secret) query(q queryFunc, prefix, key string) string {
	const me = "query"

	//
	// parse key: aws-secretsmanager:region:name:json_field
	//

	fields := strings.SplitN(key, ":", 4)
	if len(fields) < 3 {
		s.options.Printf("%s: missing fields: %s", me, key)
		return key
	}

	if fields[0] != prefix {
		s.options.Printf("%s: missing prefix='%s': %s", me, prefix, key)
		return key
	}

	region := fields[1]
	secretName := fields[2]
	var jsonField string
	if len(fields) > 3 {
		jsonField = fields[3]
	}

	s.options.Printf("%s: key='%s' json_field=%s",
		me, key, jsonField)

	//
	// retrieve secret
	//

	begin := time.Now()

	secretString, errSecret := s.retrieve(q, region, secretName, jsonField)

	s.options.Printf("%s: query: key='%s': elapsed: %v",
		me, key, time.Since(begin))

	if errSecret != nil {
		s.options.Printf("%s: secret error: key='%s': %v",
			me, key, errSecret)
		if s.options.CrashOnQueryError {
			s.options.Printf("%s: crashing on error: key='%s': %v",
				me, key, errSecret)
			os.Exit(1)
		}
		return key
	}

	if jsonField == "" {
		// return scalar (non-JSON) secret
		s.options.Printf("%s: key='%s' json_field=%s: value=%s",
			me, key, jsonField, secretString)
		return secretString
	}

	//
	// extract field from secret in JSON
	//

	value := map[string]string{}

	errJSON := yaml.Unmarshal([]byte(secretString), &value)
	if errJSON != nil {
		s.options.Printf("%s: json error: key='%s': %v",
			me, key, errJSON)
		if s.options.CrashOnQueryError {
			s.options.Printf("%s: crashing on error: key='%s': %v",
				me, key, errJSON)
			os.Exit(1)
		}
		return secretString
	}

	fieldValue := value[jsonField]

	s.options.Printf("%s: key='%s' json_field=%s: value=%s",
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
			if elapsed < s.options.CacheTTL {
				// live entry
				secretString = cached.value
				s.options.Printf("%s: from cache: %s=%s (elapsed=%s TTL=%s)",
					me, cacheKey, secretString, elapsed, s.options.CacheTTL)
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
	s.awsConfSrc.awsConfigOptions.Region = region

	value, errSecret := q(s.options.Debug, s.options.Printf, s.awsConfSrc, secretName)
	if errSecret != nil {
		s.options.Printf("%s: secret query error: %v", me, errSecret)
		return value, errSecret
	}
	secretString = value

	//
	// retrieved value from service
	//

	s.options.Printf("%s: from store: %s=%s", me, secretName, secretString)

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

type awsConfigSource struct {
	awsConfigOptions awsconfig.Options
}

func (s *awsConfigSource) get() (aws.Config, error) {
	output, err := awsconfig.AwsConfig(s.awsConfigOptions)
	return output.AwsConfig, err
}

func (s *awsConfigSource) endpointURL() string {
	return s.awsConfigOptions.EndpointURL
}

type awsConfigSolver interface {
	get() (aws.Config, error)
	endpointURL() string
}

type queryFunc func(debug bool, printf boilerplate.FuncPrintf, getAwsConfig awsConfigSolver, name string) (string, error)

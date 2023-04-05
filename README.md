[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/udhos/boilerplate/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/udhos/boilerplate)](https://goreportcard.com/report/github.com/udhos/boilerplate)
[![Go Reference](https://pkg.go.dev/badge/github.com/udhos/boilerplate.svg)](https://pkg.go.dev/github.com/udhos/boilerplate)

# boilerplate

* [envconfig](#envconfig)
  * [Usage](#usage)
    * [Create a function to load app configuration from env vars](#create-a-function-to-load-app-configuration-from-env-vars)
    * [How to define env var DB\_URI](#how-to-define-env-var-db_uri)
      * [Option 1: Literal value](#option-1-literal-value)
      * [Option 2: Retrieve scalar value from AWS Secrets Manager](#option-2-retrieve-scalar-value-from-aws-secrets-manager)
      * [Option 3: Retrieve JSON value from AWS Secrets Manager](#option-3-retrieve-json-value-from-aws-secrets-manager)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)

# envconfig

## Usage

### Create a function to load app configuration from env vars

See example function `newConfig()` below.
Or look at [examples/envconfig-example/config.go](examples/envconfig-example/config.go).

```go
import (
	"log"

	"github.com/udhos/boilerplate/awsconfig"
	"github.com/udhos/boilerplate/envconfig"
)

type appConfig struct {
	databaseURI  string
	bogus        string
	databaseCode int
	databaseTidy bool
}

func newConfig() appConfig {

	awsConfOptions := awsconfig.Options{}

	awsConf, errAwsConf := awsconfig.AwsConfig(awsConfOptions)
	if errAwsConf != nil {
		log.Printf("aws config error: %v", errAwsConf)
	}

	envOptions := envconfig.Options{
		QuerySecretsManager: true,
		QueryParameterStore: true,
		AwsConfig:           awsConf.AwsConfig,
	}

	env := envconfig.New(envOptions)

	return appConfig{
		databaseURI:  env.String("DB_URI", "http://test-db"),
		bogus:        env.String("DB_URI", "http://test-db"), // test cache
		databaseCode: env.Int("DB_CODE", 42),
		databaseTidy: env.Bool("DB_TIDY", false),
	}
}
```

### How to define env var DB_URI

#### Option 1: Literal value

    export DB_URI=http://real-db

#### Option 2: Retrieve scalar value from AWS Secrets Manager

If you prefix env var value with `aws-secretsmanager:`, the envconfig package will try to fetch it from AWS Secrets Manager.

NOTE: You can also use prefix `aws-parameterstore:` to retrieve from AWS Parameter Store.

    Format:
    export CONFIG_VAR=aws-secretsmanager:region:secret_name

    Example:
    export DB_URI=aws-secretsmanager::database_uri

    # `database_uri` is the name of the secret stored in AWS Secrets Manager
    # The secret `database_uri` could store any scalar value like: `http://real-db`

#### Option 3: Retrieve JSON value from AWS Secrets Manager

NOTE: You can also use prefix `aws-parameterstore:` to retrieve from AWS Parameter Store.

If you append ":<json_field>" to env var value, after the secret name, the package envconfig will retrieve the secret from AWS Secrets Manager and will attempt to extract that specific JSON field from the value.

    Format:
    export CONFIG_VAR=aws-secretsmanager:region:secret_name:json_field

    Example:
    export DB_URI=aws-secretsmanager::database:uri

    # `database` is the name of the secret stored in AWS Secrets Manager
    # `uri` is the name of the field to be retrieved from the JSON value
    # The secret `database` should store a JSON value like: `{"uri":"http://real-db"}`
    # In this example, the env var DB_URI will be assigned the value of the JSON field `uri`: `http://real-db`.



[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/udhos/boilerplate/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/udhos/boilerplate)](https://goreportcard.com/report/github.com/udhos/boilerplate)
[![Go Reference](https://pkg.go.dev/badge/github.com/udhos/boilerplate.svg)](https://pkg.go.dev/github.com/udhos/boilerplate)

# boilerplate

* [envconfig](#envconfig)
  * [Supported Stores](#supported-stores)
    * [DynamoDB](#dynamodb)
    * [Lambda](#lambda)
    * [HTTP](#http)
    * [Vault](#vault)
  * [Usage](#usage)
    * [Create a function to load app configuration from env vars](#create-a-function-to-load-app-configuration-from-env-vars)
    * [How to define env var DB\_URI](#how-to-define-env-var-db_uri)
      * [Option 1: Literal value](#option-1-literal-value)
      * [Option 2: Retrieve scalar value from AWS Secrets Manager](#option-2-retrieve-scalar-value-from-aws-secrets-manager)
      * [Option 3: Retrieve JSON value from AWS Secrets Manager](#option-3-retrieve-json-value-from-aws-secrets-manager)
* [References](#references)
  * [Vault](#vault-1)
    * [Curl](#curl)
    * [Example](#example)
    * [Testing AWS IAM Role](#testing-aws-iam-role)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)

# envconfig

## Supported Stores

```
aws-secretsmanager: CONFIG_VAR=aws-secretsmanager:region:secret_name[:field_name]
aws-parameterstore: CONFIG_VAR=aws-parameterstore:region:parameter_name[:field_name]
aws-s3:             CONFIG_VAR=aws-s3:region:bucket_name,object_name[:field_name]
aws-dynamodb:       CONFIG_VAR=aws-dynamodb:region:table_name,key_name,key_value,value_attr[:field_name]
aws-lambda:         CONFIG_VAR=aws-lambda:region:func_name,key_name,key_value,body_field[:field_name]
#http:              CONFIG_VAR=#http::method,proto,host,path,body_base64,token[:field_name]
vault:              CONFIG_VAR=vault::token,token-value,proto,host,port,secret_path[:field_name]
```

`:field_name` is optional. If provided, the object will be decoded as JSON/YAML and the specified field name will be extracted.

Examples:

```
export DB_URI=aws-secretsmanager:us-east-1:database:uri
export DB_URI=aws-parameterstore:us-east-1:/microservice9/mongodb:uri
export DB_URI=aws-s3:us-east-1:bucketParameters,app7/mongodb.yaml:uri
export DB_URI=aws-dynamodb:us-east-1:parameters,parameter,mongodb,value:uri
export DB_URI=aws-lambda:us-east-1:parameters,parameter,mongodb,body:uri
export DB_URI=vault::token,dev-only-token,http,localhost,8200,secret/myapp1/mongodb:uri

echo -n '{"parameter":"mongodb"}' | base64
eyJwYXJhbWV0ZXIiOiJtb25nb2RiIn0=

export DB_URI=#http::GET,https,tttt.lambda-url.us-east-1.on.aws,/,eyJwYXJhbWV0ZXIiOiJtb25nb2RiIn0=,Bearer secret:uri
```

### DynamoDB

    export DB_URI=aws-dynamodb:us-east-1:parameters,parameter,mongodb,value:uri
    #           Table: parameters
    #             Key: parameter=mongodb
    #  Attribute name: value
    # Attribute value: {"uri":"mongodb://127.0.0.1:27001/?retryWrites=false"}

### Lambda

    export DB_URI=aws-lambda:us-east-1:parameters,parameter,mongodb,body:uri
    #       Function: parameters
    #        Request: {"parameter":"mongodb"}
    # Response field: body
    #       Response: {"statusCode": 200,"body": "{\"uri\": \"mongodb://localhost:27017/?retryWrites=false\"}"}

### HTTP

    export DB_URI=#http::GET,https,tttt.lambda-url.us-east-1.on.aws,/,eyJwYXJhbWV0ZXIiOiJtb25nb2RiIn0=,Bearer secret:uri
    #   Method: GET
    # Protocol: https
    #     Host: tttt.lambda-url.us-east-1.on.aws
    #     Path: /
    #     Body: {"parameter":"mongodb"} (base64 encoded as eyJwYXJhbWV0ZXIiOiJtb25nb2RiIn0=)
    #    Token: Bearer secret
    # Response: {"uri":"mongodb://127.0.0.1:27001/?retryWrites=false"}

### Vault

    export DB_URI=vault::token,dev-only-token,http,localhost,8200,secret/myapp1/mongodb:uri
    # Auth method:  token
    # Token:        dev-only-token
    # Protocol:     http
    # Host:         localhost
    # Port:         8200
    # Secret path:  secret/myapp1/mongodb
    # Secret key:   mongodb
    # Response:     {"uri":"abc"}
    # JSON Field:   uri

## Usage

### Create a function to load app configuration from env vars

See example function `newConfig()` below.

Or look at [examples/envconfig-example/config.go](examples/envconfig-example/config.go).

```go
import (
	"github.com/udhos/boilerplate/envconfig"
)

type appConfig struct {
	databaseURI  string
	databaseCode int
	databaseTidy bool
}

func newConfig(env *envconfig.Env) appConfig {
	return appConfig{
		databaseURI:  env.String("DB_URI", "http://test-db"),
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

    Format:
    export CONFIG_VAR=aws-secretsmanager:region:secret_name

    Example:
    export DB_URI=aws-secretsmanager::database_uri

    # `database_uri` is the name of the secret stored in AWS Secrets Manager
    # The secret `database_uri` could store any scalar value like: `http://real-db`

#### Option 3: Retrieve JSON value from AWS Secrets Manager

If you append ":<json_field>" to env var value, after the secret name, the package envconfig will retrieve the secret from AWS Secrets Manager and will attempt to extract that specific JSON field from the value.

    Format:
    export CONFIG_VAR=aws-secretsmanager:region:secret_name:json_field

    Example:
    export DB_URI=aws-secretsmanager::database:uri

    # `database` is the name of the secret stored in AWS Secrets Manager
    # `uri` is the name of the field to be retrieved from the JSON value
    # The secret `database` should store a JSON value like: `{"uri":"http://real-db"}`
    # In this example, the env var DB_URI will be assigned the value of the JSON field `uri`: `http://real-db`.

# References

## Vault

Get started: https://developer.hashicorp.com/vault/docs/get-started/developer-qs

Vault server version: `hashicorp/vault:1.17.5`

Vault cli version:

```
$ vault version
Vault v1.17.5 (4d0c53e84094b8017d32b6e5b7f8142035c8837f), built 2024-08-30T15:54:57Z
```

```bash
docker run --rm -p 8200:8200 -e 'VAULT_DEV_ROOT_TOKEN_ID=dev-only-token' hashicorp/vault:1.17.5

export VAULT_ADDR=http://127.0.0.1:8200
vault login

(Enter Root Token: dev-only-token)

vault kv put secret/myapp1 mongodb='{"uri":"abc"}'

vault kv get secret/myapp1
```

### Curl

```
curl -H "X-Vault-Token: dev-only-token" http://127.0.0.1:8200/v1/secret/data/myapp1

{"request_id":"2a91f4df-64da-b585-c85b-f197869319b9","lease_id":"","renewable":false,"lease_duration":0,"data":{"data":{"mongodb":"{\"uri\":\"abc\"}"},"metadata":{"created_time":"2024-09-06T04:48:43.268293382Z","custom_metadata":null,"deletion_time":"","destroyed":false,"version":1}},"wrap_info":null,"warnings":null,"auth":null,"mount_type":"kv"}

data --> {"mongodb":"{\"uri\":\"abc\"}"}
```

### Example

`secret/myapp1` is created as
`secret/data/myapp1`, with key
`mongodb='{"uri":"abc"}`, and should be queried like
`vault::token,dev-only-token,http,localhost,8200,secret/myapp1/mongodb:uri`.

```
secret/myapp1/mongodb/uri:abc

secret  = mount
myapp1  = secret
mongodb = key
uri     = json field
```

```basH
$ vault kv get secret/myapp1
=== Secret Path ===
secret/data/myapp1

======= Metadata =======
Key                Value
---                -----
created_time       2024-09-06T04:48:43.268293382Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            1

===== Data =====
Key        Value
---        -----
mongodb    {"uri":"abc"}
```

### Testing AWS IAM Role

1. Run vault server with permission to AWS STS

```
key=$(echo $(grep aws_access_key_id ~/.aws/credentials | awk -F= '{print$2}'))
secret=$(echo $(grep aws_secret_access_key ~/.aws/credentials | awk -F= '{print$2}'))

docker run --rm -p 8200:8200 \
    -e 'VAULT_DEV_ROOT_TOKEN_ID=dev-only-token' \
    -e AWS_ACCESS_KEY_ID=$key \
    -e AWS_SECRET_ACCESS_KEY=$secret \
    hashicorp/vault:1.17.5
```

2. Use vault cli to configure the server

```
CLIENT_IAM_ROLE_ARN=... ;# fill this with client role arn

export VAULT_ADDR=http://127.0.0.1:8200

vault login

# enable aws auth
vault auth enable aws

# create policy for role dev-role-iam
vault policy write "example-policy" -<<EOF
path "secret/*" {
  capabilities = ["create", "read"]
}
EOF

# create role dev-role-iam
vault write \
  auth/aws/role/dev-role-iam \
  auth_type=iam \
  policies=example-policy \
  max_ttl=24h \
  bound_iam_principal_arn=$CLIENT_IAM_ROLE_ARN

# Put a secret to query later
vault kv put secret/myapp1 mongodb='{"uri":"abc"}'
```

3. Test client

NOTE: Vault client sdk has some limitations: (1) It does no support AWS_PROFILE. (2) It does not support credential files (`~/.aws/credentials`). It only worked with aws env vars (role credentials from env vars).

```
# Login into $CLIENT_IAM_ROLE_ARN with `aws sts assume-role` and put values into env vars.

# Then run:

export VAULT=vault::,dev-role-iam,http,localhost,8200,secret/myapp1/mongodb:uri
vault-example
```

If everything works properly, look for these lines:

```
2024/09/07 23:27:01 ##### RESULT1: vault::,dev-role-iam,http,localhost,8200,secret/myapp1/mongodb:uri: abc

2024/09/07 23:27:01 ##### RESULT2: vault::,dev-role-iam,http,localhost,8200,secret/myapp1/mongodb:uri: abc
```

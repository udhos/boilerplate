package secret

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/udhos/boilerplate/boilerplate"
	"gopkg.in/yaml.v3"
)

/*
export DB_URI=aws-lambda:us-east-1:parameters,parameter,mongodb,body:uri
#       Function: parameters
#        Request: {"parameter":"mongodb"}
# Response field: body
#       Response: {"statusCode": 200,"body": "{\"uri\": \"mongodb://localhost:27017/?retryWrites=false\"}"}
*/
func queryLambda(_ /*debug*/ bool, _ /*printf*/ boilerplate.FuncPrintf, getAwsConfig AwsConfigSolver, lambdaOptions string) (string, error) {
	const me = "queryLambda"

	options := strings.SplitN(lambdaOptions, ",", 4)
	if len(options) < 4 {
		return "", fmt.Errorf("%s: bad lambda options, expecting 4 fields - got: '%s'",
			me, lambdaOptions)
	}

	functionName := options[0]
	keyName := options[1]
	keyValue := options[2]
	responseField := options[3]

	request := fmt.Sprintf(`{"%s":"%s"}`, keyName, keyValue)

	requestBytes := []byte(request)

	awsConfig, errAwsConfig := getAwsConfig.get()
	if errAwsConfig != nil {
		return "", errAwsConfig
	}

	clientLambda := lambda.NewFromConfig(awsConfig, func(o *lambda.Options) {
		if endpoint := getAwsConfig.endpointURL(); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	input := &lambda.InvokeInput{
		FunctionName: &functionName,
		Payload:      requestBytes,
	}

	resp, errInvoke := clientLambda.Invoke(context.TODO(), input)
	if errInvoke != nil {
		return "", errInvoke
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: Invoke lambda function=%s bad status=%d payload: %s",
			me, functionName, resp.StatusCode, resp.Payload)
	}

	var funcError string
	if resp.FunctionError != nil {
		funcError = *resp.FunctionError
	}
	if funcError != "" {
		return "", fmt.Errorf("%s: Invoke lambda function=%s function_error='%s' payload: %s",
			me, functionName, funcError, resp.Payload)
	}

	payload := map[string]string{}

	errUnmarshal := yaml.Unmarshal(resp.Payload, &payload)
	if errUnmarshal != nil {
		return "", errUnmarshal
	}

	response, found := payload[responseField]
	if !found {
		return "", fmt.Errorf("%s: Invoke lambda function=%s: missing response field: '%s': %s",
			me, functionName, responseField, resp.Payload)
	}

	return response, nil
}

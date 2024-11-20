package secret

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/udhos/boilerplate/boilerplate"
)

func queryParameter(_ /*debug*/ bool, _ /*printf*/ boilerplate.FuncPrintf, getAwsConfig AwsConfigSolver, parameterName string) (string, error) {

	awsConfig, errAwsConfig := getAwsConfig.get()
	if errAwsConfig != nil {
		return "", errAwsConfig
	}

	sm := ssm.NewFromConfig(awsConfig, func(o *ssm.Options) {
		if endpoint := getAwsConfig.endpointURL(); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	input := &ssm.GetParameterInput{
		Name:           aws.String(parameterName),
		WithDecryption: aws.Bool(true),
	}

	resp, errParameter := sm.GetParameter(context.TODO(), input)

	if errParameter != nil {
		return "", errParameter
	}

	return *resp.Parameter.Value, nil
}

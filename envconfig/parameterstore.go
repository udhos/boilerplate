// Package envconfig loads configuration from env vars.
package envconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func queryParameter(awsConfig aws.Config, parameterName string) (string, error) {
	const me = "queryParameter"

	sm := ssm.NewFromConfig(awsConfig)

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

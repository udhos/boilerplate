package secret

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func queryParameter(getAwsConfig awsConfigSolver, parameterName string) (string, error) {
	const me = "queryParameter"

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

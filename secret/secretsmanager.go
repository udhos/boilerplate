package secret

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/udhos/boilerplate/boilerplate"
)

func querySecret(_ /*debug*/ bool, _ /*printf*/ boilerplate.FuncPrintf, getAwsConfig AwsConfigSolver, secretName string) (string, error) {

	awsConfig, errAwsConfig := getAwsConfig.get()
	if errAwsConfig != nil {
		return "", errAwsConfig
	}

	sm := secretsmanager.NewFromConfig(awsConfig, func(o *secretsmanager.Options) {
		if endpoint := getAwsConfig.endpointURL(); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

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

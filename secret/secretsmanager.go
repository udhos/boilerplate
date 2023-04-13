package secret

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func querySecret(getAwsConfig awsConfigSolver, secretName string) (string, error) {
	const me = "querySecret"

	awsConfig, errAwsConfig := getAwsConfig.get()
	if errAwsConfig != nil {
		return "", errAwsConfig
	}

	sm := secretsmanager.NewFromConfig(awsConfig)

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

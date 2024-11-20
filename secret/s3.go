package secret

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/udhos/boilerplate/boilerplate"
)

func queryS3(_ /*debug*/ bool, _ /*printf*/ boilerplate.FuncPrintf, getAwsConfig AwsConfigSolver, bucketAndKey string) (string, error) {
	const me = "queryS3"

	bucketName, objectKey, found := strings.Cut(bucketAndKey, ",")
	if !found {
		return "", fmt.Errorf("%s: bad bucket object, expecting 'bucket,key' - got: '%s'",
			me, bucketAndKey)
	}

	awsConfig, errAwsConfig := getAwsConfig.get()
	if errAwsConfig != nil {
		return "", errAwsConfig
	}

	s3client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		if endpoint := getAwsConfig.endpointURL(); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}

	result, errS3 := s3client.GetObject(context.TODO(), input)
	if errS3 != nil {
		return "", errS3
	}

	body, err := io.ReadAll(result.Body)

	bodyStr := string(body)

	return bodyStr, err
}

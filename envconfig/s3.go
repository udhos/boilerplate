// Package envconfig loads configuration from env vars.
package envconfig

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func queryS3(awsConfig aws.Config, bucketAndKey string) (string, error) {
	const me = "queryS3"

	bucketName, objectKey, found := strings.Cut(bucketAndKey, ",")
	if !found {
		return "", fmt.Errorf("%s: bad bucket object, expecting 'bucket,key' - got: '%s'",
			me, bucketAndKey)
	}

	s3client := s3.NewFromConfig(awsConfig)

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

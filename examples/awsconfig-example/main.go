// Package main implements a example application for awsconfig package.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/udhos/boilerplate/awsconfig"
)

func main() {
	options := awsconfig.Options{}
	awsCfg, errCfg := awsconfig.AwsConfig(options)
	if errCfg != nil {
		log.Printf("could not get aws config: %v", errCfg)
	}

	fmt.Printf("STS account ID: %s\n", awsCfg.StsAccountID)
	fmt.Printf("STS ARN: %s\n", awsCfg.StsArn)
	fmt.Printf("STS UserId: %s\n", awsCfg.StsUserID)

	creds, errCreds := awsCfg.AwsConfig.Credentials.Retrieve(context.TODO())
	if errCreds != nil {
		log.Printf("could not get aws credentials: %v", errCreds)
	}

	fmt.Printf("aws access key ID: %s\n", creds.AccessKeyID)
}

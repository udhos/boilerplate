// Package main implements a example application for awsconfig package.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/udhos/boilerplate/awsconfig"
	"github.com/udhos/boilerplate/boilerplate"
)

func main() {
	me := filepath.Base(os.Args[0])
	log.Println(boilerplate.LongVersion(me))

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

// Package awsconfig loads AWS configuration.
package awsconfig

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Options provide optional parameters for AwsConfig.
type Options struct {
	Region          string
	RoleArn         string
	RoleSessionName string
	Printf          FuncPrintf // defaults to log.Printf
}

// FuncPrintf is a helper type for logging function.
type FuncPrintf func(format string, v ...any)

// AwsConfig provides a configuration to initialize clients for AWS services.
// If roleArn is provided, it assumes the role.
// Otherwise it works with default credentials.
func AwsConfig(opt Options) (aws.Config, error) {
	const me = "awsConfig"

	if opt.Printf == nil {
		opt.Printf = log.Printf
	}

	cfg, errConfig := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(opt.Region))
	if errConfig != nil {
		opt.Printf("%s: load config: %v", me, errConfig)
		return cfg, errConfig
	}

	if opt.RoleArn != "" {
		//
		// AssumeRole
		//
		opt.Printf("%s: AssumeRole: arn: %s", me, opt.RoleArn)
		clientSts := sts.NewFromConfig(cfg)
		cfg2, errConfig2 := config.LoadDefaultConfig(
			context.TODO(), config.WithRegion(opt.Region),
			config.WithCredentialsProvider(aws.NewCredentialsCache(
				stscreds.NewAssumeRoleProvider(
					clientSts,
					opt.RoleArn,
					func(o *stscreds.AssumeRoleOptions) {
						o.RoleSessionName = opt.RoleSessionName
					},
				)),
			),
		)
		if errConfig2 != nil {
			opt.Printf("%s: AssumeRole %s: error: %v", me, opt.RoleArn, errConfig2)
			return cfg, errConfig
		}
		cfg = cfg2
	}

	{
		// show caller identity
		clientSts := sts.NewFromConfig(cfg)
		input := sts.GetCallerIdentityInput{}
		respSts, errSts := clientSts.GetCallerIdentity(context.TODO(), &input)
		if errSts != nil {
			opt.Printf("%s: GetCallerIdentity: error: %v", me, errSts)
		} else {
			opt.Printf("%s: GetCallerIdentity: Account=%s ARN=%s UserId=%s", me, *respSts.Account, *respSts.Arn, *respSts.UserId)
		}
	}

	return cfg, nil
}

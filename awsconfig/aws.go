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
	RoleExternalID  string
	Printf          FuncPrintf // defaults to log.Printf
}

// Output holds returned result.
type Output struct {
	AwsConfig    aws.Config // AwsConfig holds the desired configuration.
	StsAccountID string
	StsArn       string
	StsUserID    string
}

// FuncPrintf is a helper type for logging function.
type FuncPrintf func(format string, v ...any)

// AwsConfig provides a configuration to initialize clients for AWS services.
// If roleArn is provided, it assumes the role.
// Otherwise it works with default credentials.
func AwsConfig(opt Options) (Output, error) {
	const me = "AwsConfig"

	var out Output

	if opt.Printf == nil {
		opt.Printf = log.Printf
	}

	cfg, errConfig := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(opt.Region))
	if errConfig != nil {
		opt.Printf("%s: load config: %v", me, errConfig)
		return out, errConfig
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
						if opt.RoleExternalID != "" {
							o.ExternalID = aws.String(opt.RoleExternalID)
						}
					},
				)),
			),
		)
		if errConfig2 != nil {
			opt.Printf("%s: AssumeRole %s: error: %v", me, opt.RoleArn, errConfig2)
			return out, errConfig
		}
		cfg = cfg2
	}

	out.AwsConfig = cfg

	{
		// show caller identity
		clientSts := sts.NewFromConfig(cfg)
		input := sts.GetCallerIdentityInput{}
		respSts, errSts := clientSts.GetCallerIdentity(context.TODO(), &input)
		if errSts != nil {
			opt.Printf("%s: GetCallerIdentity: error: %v", me, errSts)
		} else {
			out.StsAccountID = *respSts.Account
			out.StsArn = *respSts.Arn
			out.StsUserID = *respSts.UserId
			opt.Printf("%s: GetCallerIdentity: Account=%s ARN=%s UserId=%s",
				me, out.StsAccountID, out.StsArn, out.StsUserID)
		}
	}

	return out, nil
}

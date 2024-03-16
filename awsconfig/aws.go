// Package awsconfig loads AWS configuration.
package awsconfig

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/udhos/boilerplate/boilerplate"
)

// Options provide optional parameters for AwsConfig.
type Options struct {
	Region               string
	RoleArn              string
	RoleSessionName      string
	RoleExternalID       string
	EndpointURL          string
	Printf               boilerplate.FuncPrintf // defaults to log.Printf
	RetryMaxAttempts     int
	RetryMaxBackoffDelay time.Duration
}

// Output holds returned result.
type Output struct {
	AwsConfig    aws.Config // AwsConfig holds the desired configuration.
	StsAccountID string
	StsArn       string
	StsUserID    string
}

// AwsConfig provides a configuration to initialize clients for AWS services.
// If roleArn is provided, it assumes the role.
// Otherwise it works with default credentials.
func AwsConfig(opt Options) (Output, error) {
	const me = "AwsConfig"

	var out Output

	if opt.Printf == nil {
		opt.Printf = log.Printf
	}
	if opt.RetryMaxAttempts == 0 {
		opt.RetryMaxAttempts = 6 // increase from default=3 to 6
	}
	if opt.RetryMaxBackoffDelay == 0 {
		opt.RetryMaxBackoffDelay = 40 * time.Second // increase from default=20 to 40
	}

	var cfg aws.Config
	var errConfig error

	optionsFunc := config.WithRetryer(func() aws.Retryer {
		var r aws.Retryer
		r = retry.NewStandard()
		r = retry.AddWithMaxAttempts(r, opt.RetryMaxAttempts)
		return retry.AddWithMaxBackoffDelay(r, opt.RetryMaxBackoffDelay)
	})

	if opt.EndpointURL == "" {
		cfg, errConfig = config.LoadDefaultConfig(context.TODO(),
			optionsFunc, config.WithRegion(opt.Region))
	} else {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string,
			options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           opt.EndpointURL,
				SigningRegion: opt.Region,
			}, nil
		})
		cfg, errConfig = config.LoadDefaultConfig(context.TODO(),
			optionsFunc, config.WithEndpointResolverWithOptions(customResolver))
	}

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

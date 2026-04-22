/*
Package aws provides wrappers and helpers for interacting with AWS services
using the official AWS SDK v2 for Go.
*/
package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// GetConfig returns an AWS configuration based on the provided profile and region.
func GetConfig(profile, region string) (aws.Config, error) {
	return config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region),
	)
}

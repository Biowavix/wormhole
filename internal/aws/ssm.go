/*
Package aws provides wrappers and helpers for interacting with AWS services.
This file specifically handles SSM Parameter Store calls.
*/
package aws

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// GetRDSTags retrieves tags for a given RDS instance ARN.
func GetRDSTags(cfg aws.Config, arn string) (map[string]string, error) {
	client := rds.NewFromConfig(cfg)
	input := &rds.ListTagsForResourceInput{
		ResourceName: aws.String(arn),
	}
	output, err := client.ListTagsForResource(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	
	tags := make(map[string]string)
	for _, t := range output.TagList {
		tags[*t.Key] = *t.Value
	}
	return tags, nil
}

// GetParameter retrieves the value of an SSM parameter.
func GetParameter(cfg aws.Config, name string, decrypt bool) (string, error) {
	client := ssm.NewFromConfig(cfg)
	
	input := &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(decrypt),
	}
	output, err := client.GetParameter(context.TODO(), input)
	if err != nil {
		return "", err
	}
	
	return *output.Parameter.Value, nil
}

// ParameterExists checks if an SSM parameter exists.
func ParameterExists(cfg aws.Config, name string) bool {
	_, err := GetParameter(cfg, name, false)
	return err == nil
}

// FindBestParameterPath searches for the most likely credential path for an RDS instance.
func FindBestParameterPath(cfg aws.Config, prefix, rdsName string) (string, error) {
	client := ssm.NewFromConfig(cfg)
	
	// Use GetParametersByPath for more targeted search
	input := &ssm.GetParametersByPathInput{
		Path:           aws.String("/" + prefix + "/"),
		Recursive:      aws.Bool(true),
		WithDecryption: aws.Bool(false),
	}
	
	var bestPath string
	maxScore := -1
	
	paginator := ssm.NewGetParametersByPathPaginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			// If prefix path doesn't exist, try just a broad contains search as fallback
			return findFuzzyFallback(cfg, prefix, rdsName)
		}
		
		for _, p := range page.Parameters {
			name := *p.Name
			if !strings.Contains(name, "/db/") {
				continue
			}
			
			idx := strings.LastIndex(name, "/db/")
			path := name[:idx+4]
			
			score := calculatePathScore(path, prefix, rdsName)
			if score > maxScore {
				maxScore = score
				bestPath = path
			}
		}
	}
	
	if bestPath == "" {
		return findFuzzyFallback(cfg, prefix, rdsName)
	}
	
	return bestPath, nil
}

func calculatePathScore(path, prefix, rdsName string) int {
	score := 0
	pathLower := strings.ToLower(path)
	rdsNameLower := strings.ToLower(rdsName)
	rdsParts := strings.Split(rdsNameLower, "-")
	
	for _, part := range rdsParts {
		if part == "rds" || part == prefix || len(part) < 3 {
			continue
		}
		if strings.Contains(pathLower, part) {
			score += 20 // Increased weight for part matches
		} else {
			score -= 5 // Penalty for missing a core part of the name
		}
	}
	
	// Huge bonus for exact rdsName match in the path
	if strings.Contains(pathLower, rdsNameLower) {
		score += 100
	}
	
	// Reward shorter paths that are still specific (Ockham's razor)
	score -= len(path) / 2
	
	return score
}

func findFuzzyFallback(cfg aws.Config, prefix, rdsName string) (string, error) {
	client := ssm.NewFromConfig(cfg)
	input := &ssm.DescribeParametersInput{
		ParameterFilters: []ssmtypes.ParameterStringFilter{
			{
				Key:    aws.String("Name"),
				Option: aws.String("Contains"),
				Values: []string{rdsName},
			},
		},
	}
	
	output, err := client.DescribeParameters(context.TODO(), input)
	if err != nil || len(output.Parameters) == 0 {
		return "", nil
	}
	
	name := *output.Parameters[0].Name
	idx := strings.LastIndex(name, "/db/")
	if idx == -1 {
		return "", nil
	}
	return name[:idx+4], nil
}

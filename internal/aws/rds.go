/*
Package aws provides wrappers and helpers for interacting with AWS services.
This file specifically handles RDS service calls.
*/
package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
)

// ListRDSInstances retrieves all RDS instance identifiers, optionally filtered.
func ListRDSInstances(cfg aws.Config, filter string) ([]string, error) {
	client := rds.NewFromConfig(cfg)
	
	input := &rds.DescribeDBInstancesInput{}
	output, err := client.DescribeDBInstances(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	
	var instances []string
	for _, db := range output.DBInstances {
		name := *db.DBInstanceIdentifier
		if filter == "" || strings.Contains(strings.ToLower(name), strings.ToLower(filter)) {
			instances = append(instances, name)
		}
	}
	
	return instances, nil
}

// DBEndpoint represents the host and port of a database.
type DBEndpoint struct {
	Address string
	Port    int32
}

// GetRDSInstance returns the full RDS instance object.
func GetRDSInstance(cfg aws.Config, identifier string) (*types.DBInstance, error) {
	client := rds.NewFromConfig(cfg)
	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(identifier),
	}
	output, err := client.DescribeDBInstances(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	if len(output.DBInstances) == 0 {
		return nil, fmt.Errorf("instance not found")
	}
	return &output.DBInstances[0], nil
}

// GetRDSEndpoint retrieves the connection endpoint for a specific RDS instance.
func GetRDSEndpoint(cfg aws.Config, rdsName string) (*DBEndpoint, error) {
	db, err := GetRDSInstance(cfg, rdsName)
	if err != nil {
		return nil, err
	}
	
	return &DBEndpoint{
		Address: *db.Endpoint.Address,
		Port:    *db.Endpoint.Port,
	}, nil
}

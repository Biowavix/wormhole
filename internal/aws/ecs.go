/*
Package aws provides wrappers and helpers for interacting with AWS services.
This file specifically handles ECS service calls.
*/
package aws

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

// ListClusters retrieves all ECS cluster names, optionally filtered by a search string.
func ListClusters(cfg aws.Config, filter string) ([]string, error) {
	client := ecs.NewFromConfig(cfg)
	
	paginator := ecs.NewListClustersPaginator(client, &ecs.ListClustersInput{})
	
	var clusters []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		
		for _, arn := range page.ClusterArns {
			// Extract name from ARN (arn:aws:ecs:region:account:cluster/name)
			parts := strings.Split(arn, "/")
			name := parts[len(parts)-1]
			
			if filter == "" || strings.Contains(strings.ToLower(name), strings.ToLower(filter)) {
				clusters = append(clusters, name)
			}
		}
	}
	
	return clusters, nil
}

// GetClusterInstance retrieves the first active EC2 instance ID for a cluster.
func GetClusterInstance(cfg aws.Config, clusterName string) (string, error) {
	client := ecs.NewFromConfig(cfg)
	
	listInput := &ecs.ListContainerInstancesInput{
		Cluster: aws.String(clusterName),
	}
	listOutput, err := client.ListContainerInstances(context.TODO(), listInput)
	if err != nil {
		return "", err
	}
	
	if len(listOutput.ContainerInstanceArns) == 0 {
		return "", nil
	}
	
	descInput := &ecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(clusterName),
		ContainerInstances: listOutput.ContainerInstanceArns,
	}
	descOutput, err := client.DescribeContainerInstances(context.TODO(), descInput)
	if err != nil {
		return "", err
	}
	
	if len(descOutput.ContainerInstances) == 0 {
		return "", nil
	}
	
	return *descOutput.ContainerInstances[0].Ec2InstanceId, nil
}

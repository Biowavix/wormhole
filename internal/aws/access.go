/*
Package aws provides wrappers and helpers for interacting with AWS services.
This file handles security group and network access diagnostics.
*/
package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// RDSAccessInfo contains details about what has access to an RDS.
type RDSAccessInfo struct {
	RDSName     string
	AllowedSGs  []string
	AllowedIPs  []string
	Instances   []InstanceInfo
}

type InstanceInfo struct {
	ID        string
	Name      string
	PrivateIP string
	SGID      string
}

// CheckRDSAccess identifies which EC2 instances have security group access to an RDS on port 5432.
func CheckRDSAccess(cfg aws.Config, rdsName string) (*RDSAccessInfo, error) {
	rdsClient := rds.NewFromConfig(cfg)
	ec2Client := ec2.NewFromConfig(cfg)

	// 1. Get RDS Security Groups
	rdsOut, err := rdsClient.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(rdsName),
	})
	if err != nil || len(rdsOut.DBInstances) == 0 {
		return nil, fmt.Errorf("failed to describe RDS: %v", err)
	}

	rdsSGs := rdsOut.DBInstances[0].VpcSecurityGroups
	var allowedSourceSGs []string
	var allowedIPs []string

	// 2. Inspect Ingress Rules for each RDS SG
	for _, rsg := range rdsSGs {
		sgOut, err := ec2Client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{
			GroupIds: []string{*rsg.VpcSecurityGroupId},
		})
		if err != nil {
			continue
		}

		for _, sg := range sgOut.SecurityGroups {
			for _, perm := range sg.IpPermissions {
				// Check if port 5432 is covered
				if isPortAllowed(perm, 5432) {
					for _, pair := range perm.UserIdGroupPairs {
						allowedSourceSGs = append(allowedSourceSGs, *pair.GroupId)
					}
					for _, ip := range perm.IpRanges {
						allowedIPs = append(allowedIPs, *ip.CidrIp)
					}
				}
			}
		}
	}

	// 3. Find instances in those source SGs
	var instances []InstanceInfo
	if len(allowedSourceSGs) > 0 {
		instOut, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("instance.group-id"),
					Values: allowedSourceSGs,
				},
				{
					Name:   aws.String("instance-state-name"),
					Values: []string{"running"},
				},
			},
		})
		if err == nil {
			for _, res := range instOut.Reservations {
				for _, inst := range res.Instances {
					name := ""
					for _, t := range inst.Tags {
						if *t.Key == "Name" {
							name = *t.Value
						}
					}
					instances = append(instances, InstanceInfo{
						ID:        *inst.InstanceId,
						Name:      name,
						PrivateIP: *inst.PrivateIpAddress,
						SGID:      (*inst.SecurityGroups[0].GroupId),
					})
				}
			}
		}
	}

	return &RDSAccessInfo{
		RDSName:    rdsName,
		AllowedSGs: allowedSourceSGs,
		AllowedIPs: allowedIPs,
		Instances:  instances,
	}, nil
}

func isPortAllowed(perm types.IpPermission, port int32) bool {
	if perm.IpProtocol != nil && *perm.IpProtocol == "-1" {
		return true // All protocols
	}
	if perm.FromPort == nil || perm.ToPort == nil {
		return false
	}
	return port >= *perm.FromPort && port <= *perm.ToPort
}

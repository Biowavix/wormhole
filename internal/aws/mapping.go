/*
Package aws provides wrappers and helpers for interacting with AWS services.
This file contains the manual mapping between RDS instances and Parameter Store paths.
*/
package aws

// RDSMapping contains manual overrides for RDS instance credential paths.
var RDSMapping = map[string]string{
	"boidas-tenants-rds":      "/boidas/tenants-rds/db/",
	"veriadmin-biprod-rds":    "/veriadmin/biprod-rds/db/",
	"veriadmin-common-rds":    "/veriadmin/common-rds/db/",
	"veriadmin-rds":           "/veriadmin/db/",
	"verisaas2-altan-web-rds": "/verisaas2/altan-web-rds/db/",
	"verisaas2-common-rds":    "/verisaas2/common-rds/db/",
	"verisaas2-eks-rds":       "/verisaas2/eks-rds/db/",
	"verisaas2-phy-iam-rds":   "/verisaas2/phy-iam-rds/db/",
	"verisaas2-rds":           "/verisaas2/db/",
	"verisaas2-rds2":          "UNKNOWN", // Por favor, rellena este
	"verisaas2-squid-rds":     "/verisaas2/squid-rds/db/",
	"verisaas2-vcsp-rds":      "/verisaas2/vcsp-rds/db/",
	"verisaas2-voice-rds":     "/verisaas2/voice-rds/db/",
	"vs2-play-eu-west-1":      "UNKNOWN", // Por favor, rellena este
}

// RDSBridgeMapping contains manual overrides for ECS clusters used as bridges.
var RDSBridgeMapping = map[string]string{
	"veriadmin-rds":           "veriadmin-idmanager-cluster",
	"veriadmin-common-rds":    "veriadmin-idmanager-cluster",
	"veriadmin-biprod-rds":    "veriadmin-idmanager-cluster",
	"verisaas2-common-rds":    "verisaas2-democenter-FE-cluster",
	"verisaas2-rds":           "verisaas2-vaas-cluster",
	"boidas-tenants-rds":      "boidas-default-cluster",
}

// GetManualPath returns the manual path override for an RDS instance if it exists.
func GetManualPath(rdsName string) (string, bool) {
	path, ok := RDSMapping[rdsName]
	return path, ok
}

// GetManualBridge returns the manual bridge cluster override for an RDS instance if it exists.
func GetManualBridge(rdsName string) (string, bool) {
	cluster, ok := RDSBridgeMapping[rdsName]
	return cluster, ok
}

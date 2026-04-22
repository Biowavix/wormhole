/*
Package utils provides helper functions for system-level operations.
This file handles launching the SSM session.
*/
package utils

import (
	"fmt"
	"os"
	"os/exec"
)

// StartSSMSession launches an interactive SSM session to the target instance.
func StartSSMSession(profile, region, instanceID string) error {
	cmd := exec.Command("aws", "ssm", "start-session",
		"--target", instanceID,
		"--profile", profile,
		"--region", region,
	)
	
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// StartSSMTunnel starts an SSM port forwarding session in the background.
func StartSSMTunnel(profile, region, targetInstance, remoteHost string, remotePort, localPort int32) (*exec.Cmd, error) {
	params := fmt.Sprintf(`{"host":["%s"],"portNumber":["%d"],"localPortNumber":["%d"]}`, remoteHost, remotePort, localPort)
	
	cmd := exec.Command("aws", "ssm", "start-session",
		"--target", targetInstance,
		"--document-name", "AWS-StartPortForwardingSessionToRemoteHost",
		"--parameters", params,
		"--profile", profile,
		"--region", region,
	)
	
	err := cmd.Start()
	return cmd, err
}

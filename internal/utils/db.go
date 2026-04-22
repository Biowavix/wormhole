/*
Package utils provides helper functions for system-level operations.
This file handles database client calls.
*/
package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunPsql launches psql with the provided credentials and host/port.
func RunPsql(user, password, dbname, host string, port int32) error {
	if host == "localhost" {
		host = "127.0.0.1"
	}
	cmd := exec.Command("psql", "-h", host, "-p", fmt.Sprintf("%d", port), "-U", user, "-d", dbname)
	
	// Create a clean environment without any pre-existing PGPASSWORD
	env := os.Environ()
	newEnv := make([]string, 0, len(env)+1)
	for _, e := range env {
		if !strings.HasPrefix(e, "PGPASSWORD=") {
			newEnv = append(newEnv, e)
		}
	}
	newEnv = append(newEnv, "PGPASSWORD="+password)
	cmd.Env = newEnv
	
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

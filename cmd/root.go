/*
Package cmd handles the CLI command definitions using Cobra.
root.go defines the base command and global flags like profile and region.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	profile string
	region  string
)

var rootCmd = &cobra.Command{
	Use:   "wh",
	Short: "Wormhole is a secure bridge for Cloud Operations",
	Long:  `A unified tool to tunnel into private infrastructure (ECS, RDS) using secure SSM bridges.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "veridas-play-ireland", "AWS profile to use")
	rootCmd.PersistentFlags().StringVar(&region, "region", "eu-west-1", "AWS region to use")
}

/*
Package cmd handles the ecs subcommands.
This file defines 'wh ecs list' and 'wh ecs connect'.
*/
package cmd

import (
	"fmt"
	"wormhole/internal/aws"
	"wormhole/internal/ui"
	"wormhole/internal/utils"
	"github.com/spf13/cobra"
)

var ecsCmd = &cobra.Command{
	Use:   "ecs",
	Short: "Manage ECS clusters and instances",
}

var ecsListCmd = &cobra.Command{
	Use:   "ls [filter]",
	Aliases: []string{"list"},
	Short: "List ECS clusters",
	Run: func(cmd *cobra.Command, args []string) {
		filter := ""
		if len(args) > 0 {
			filter = args[0]
		}

		cfg, err := aws.GetConfig(profile, region)
		if err != nil {
			fmt.Printf("Error loading AWS config: %v\n", err)
			return
		}

		clusters, err := aws.ListClusters(cfg, filter)
		if err != nil {
			fmt.Printf("Error listing clusters: %v\n", err)
			return
		}

		if len(clusters) == 0 {
			fmt.Println("No clusters found.")
			return
		}

		ui.PrintList(fmt.Sprintf("ECS Clusters (%s)", profile), clusters)
	},
}

var ecsConnCmd = &cobra.Command{
	Use:   "conn [filter]",
	Aliases: []string{"connect"},
	Short: "Connect to an ECS instance via SSM",
	Run: func(cmd *cobra.Command, args []string) {
		filter := ""
		if len(args) > 0 {
			filter = args[0]
		}

		cfg, err := aws.GetConfig(profile, region)
		if err != nil {
			fmt.Printf("Error loading AWS config: %v\n", err)
			return
		}

		clusters, err := aws.ListClusters(cfg, filter)
		if err != nil {
			fmt.Printf("Error listing clusters: %v\n", err)
			return
		}

		if len(clusters) == 0 {
			fmt.Printf("No clusters found matching '%s'\n", filter)
			return
		}

		var selectedCluster string
		if len(clusters) == 1 {
			selectedCluster = clusters[0]
			fmt.Printf("Using cluster: %s\n", selectedCluster)
		} else {
			selectedCluster, err = ui.SelectOne("Select Cluster", clusters)
			if err != nil {
				fmt.Printf("Selection error: %v\n", err)
				return
			}
		}

		if selectedCluster == "" {
			return
		}

		fmt.Printf("Finding instance in %s...\n", selectedCluster)
		instanceID, err := aws.GetClusterInstance(cfg, selectedCluster)
		if err != nil {
			fmt.Printf("Error finding instance: %v\n", err)
			return
		}

		if instanceID == "" {
			fmt.Println("No active instances found in this cluster.")
			return
		}

		fmt.Printf("Connecting to %s...\n", instanceID)
		err = utils.StartSSMSession(profile, region, instanceID)
		if err != nil {
			fmt.Printf("Session error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(ecsCmd)
	ecsCmd.AddCommand(ecsListCmd)
	ecsCmd.AddCommand(ecsConnCmd)
}

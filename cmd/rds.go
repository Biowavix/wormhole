/*
Package cmd handles the rds subcommands.
This file defines 'vops rds list' and 'vops rds connect'.
*/
package cmd

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"vops/internal/aws"
	"vops/internal/ui"
	"vops/internal/utils"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFD1"))
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	keyStyle   = lipgloss.NewStyle().Width(15).Foreground(lipgloss.Color("#FFA500"))
	valStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
)

func printConnectionSummary(rdsName, endpoint, path, user, pass, bridge string, lPort, rPort int32) {
	fmt.Println("\n" + titleStyle.Render("─── CONNECTION PARAMETERS ───"))
	fmt.Printf("%s %s\n", keyStyle.Render("RDS Target:"), valStyle.Render(rdsName))
	fmt.Printf("%s %s:%d\n", keyStyle.Render("Endpoint:"), valStyle.Render(endpoint), rPort)
	fmt.Printf("%s %s\n", keyStyle.Render("Metadata Path:"), infoStyle.Render(path))
	fmt.Printf("%s %s\n", keyStyle.Render("Credentials:"), valStyle.Render(fmt.Sprintf("%s / %s...", user, pass[:3])))
	fmt.Printf("%s %s\n", keyStyle.Render("Bridge:"), valStyle.Render(bridge))
	fmt.Printf("%s %s\n", keyStyle.Render("Local Port:"), valStyle.Render(fmt.Sprintf("%d", lPort)))
	fmt.Println(titleStyle.Render("─────────────────────────────") + "\n")
}

var rdsCmd = &cobra.Command{
	Use:   "rds",
	Short: "Manage RDS instances",
}

var rdsListCmd = &cobra.Command{
	Use:   "ls [filter]",
	Aliases: []string{"list"},
	Short: "List RDS instances",
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

		instances, err := aws.ListRDSInstances(cfg, filter)
		if err != nil {
			fmt.Printf("Error listing instances: %v\n", err)
			return
		}

		if len(instances) == 0 {
			fmt.Println("No RDS instances found.")
			return
		}

		ui.PrintList(fmt.Sprintf("RDS Instances (%s)", profile), instances)
	},
}

var rdsConnCmd = &cobra.Command{
	Use:   "conn [filter]",
	Aliases: []string{"connect"},
	Short: "Connect to an RDS database via SSM tunnel",
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

		instances, err := aws.ListRDSInstances(cfg, filter)
		if err != nil {
			fmt.Printf("Error listing RDS: %v\n", err)
			return
		}

		if len(instances) == 0 {
			fmt.Printf("No RDS found matching '%s'\n", filter)
			return
		}

		var selectedRDS string
		if len(instances) == 1 {
			selectedRDS = instances[0]
			fmt.Printf("Using RDS: %s\n", selectedRDS)
		} else {
			selectedRDS, err = ui.SelectOne("Select RDS", instances)
			if err != nil {
				fmt.Printf("Selection error: %v\n", err)
				return
			}
		}

		if selectedRDS == "" {
			return
		}

		selectedInstance, err := aws.GetRDSInstance(cfg, selectedRDS)
		if err != nil {
			fmt.Printf("Error getting instance details: %v\n", err)
			return
		}

		endpoint := &aws.DBEndpoint{
			Address: *selectedInstance.Endpoint.Address,
			Port:    *selectedInstance.Endpoint.Port,
		}

		prefix := strings.Split(selectedRDS, "-")[0]
		suffix := strings.TrimPrefix(selectedRDS, prefix+"-")

		// --- Metadata Discovery (The "State of the Art" Way) ---
		fmt.Println("Discovering metadata for connection...")
		tags, _ := aws.GetRDSTags(cfg, *selectedInstance.DBInstanceArn)
		
		var foundPath string
		var manualBridgeOverride string

		// 1. Discover Credential Path
		if tagPath, ok := tags["vops:credential-path"]; ok {
			fmt.Println("✓ Found credential path via resource tags.")
			foundPath = tagPath
		} else if mPath, ok := aws.GetManualPath(selectedRDS); ok {
			fmt.Println("✓ Using manual mapping for credentials.")
			foundPath = mPath
		} else {
			fmt.Println("? No tags or mapping found. Using fuzzy search...")
			var err error
			foundPath, err = aws.FindBestParameterPath(cfg, prefix, selectedRDS)
			if err != nil || foundPath == "" {
				fmt.Println("Error: Could not find any parameter path for credentials.")
				return
			}
		}

		// 2. Discover Bridge Cluster
		manualBridgeOverride = ""
		if tagBridge, ok := tags["vops:bridge-cluster"]; ok {
			fmt.Printf("✓ Found bridge cluster via resource tags: %s\n", tagBridge)
			manualBridgeOverride = tagBridge
		} else if mBridge, ok := aws.GetManualBridge(selectedRDS); ok {
			fmt.Printf("✓ Using manual bridge mapping: %s\n", mBridge)
			manualBridgeOverride = mBridge
		}

		fmt.Printf("Using credentials path: %s\n", foundPath)
		dbUser, err := aws.GetParameter(cfg, foundPath+"user", false)
		if err != nil {
			fmt.Printf("Error fetching DB user: %v\n", err)
			return
		}
		dbPass, err := aws.GetParameter(cfg, foundPath+"password", true)
		if err != nil {
			fmt.Printf("Error fetching DB password: %v\n", err)
			return
		}

		// Clean credentials and endpoint
		endpoint.Address = strings.TrimSpace(endpoint.Address)
		dbUser = strings.TrimSpace(dbUser)
		dbPass = strings.TrimSpace(dbPass)

		maskedPass := ""
		if len(dbPass) > 3 {
			maskedPass = dbPass[:3] + "..."
		}
		fmt.Printf("Connecting as %s (Password: %s)\n", dbUser, maskedPass)

		dbName, _ := aws.GetParameter(cfg, foundPath+"default_name", false)
		if dbName == "" {
			dbName = "postgres"
		}
		dbName = strings.TrimSpace(dbName)

		// Bridge discovery (Manual override or Robust loop)
		fmt.Println("Finding bridge instance...")
		var bridgeCluster string
		var instanceID string

		// 1. Check manual override
		if manualBridgeOverride != "" {
			id, err := aws.GetClusterInstance(cfg, manualBridgeOverride)
			if err == nil && id != "" {
				bridgeCluster = manualBridgeOverride
				instanceID = id
			} else {
				fmt.Printf("Manual bridge %s has no active instances, falling back...\n", manualBridgeOverride)
			}
		}

		// 2. Fallback to robust loop if no manual bridge found or it has no instances
		if instanceID == "" {
			middlePart := strings.Split(suffix, "-")[0]
			searchTerms := []string{
				fmt.Sprintf("%s-%s", prefix, middlePart),
				prefix,
			}

			for _, term := range searchTerms {
				allClusters, _ := aws.ListClusters(cfg, term)
				for _, cluster := range allClusters {
					id, err := aws.GetClusterInstance(cfg, cluster)
					if err == nil && id != "" {
						bridgeCluster = cluster
						instanceID = id
						break
					}
				}
				if instanceID != "" {
					break
				}
			}
		}

		// Fallback to any cluster
		if instanceID == "" {
			fmt.Println("No matching clusters have active instances. Trying any available cluster...")
			all, _ := aws.ListClusters(cfg, "")
			for _, cluster := range all {
				id, err := aws.GetClusterInstance(cfg, cluster)
				if err == nil && id != "" {
					bridgeCluster = cluster
					instanceID = id
					break
				}
			}
		}

		if instanceID == "" {
			fmt.Println("Error: No active ECS instance found in the profile to use as a bridge.")
			return
		}

		// Tunnel and Connect
		localPort := int32(5440 + rand.Intn(500))
		
		printConnectionSummary(selectedRDS, endpoint.Address, foundPath, dbUser, dbPass, bridgeCluster, localPort, endpoint.Port)

		fmt.Printf("🚀 Establishing tunnel via %s (%s)...\n", bridgeCluster, instanceID)
		tunnelCmd, err := utils.StartSSMTunnel(profile, region, instanceID, endpoint.Address, endpoint.Port, localPort)
		if err != nil {
			fmt.Printf("Error starting tunnel: %v\n", err)
			return
		}
		defer func() {
			if tunnelCmd.Process != nil {
				tunnelCmd.Process.Kill()
			}
		}()

		fmt.Printf("⏳ Waiting for tunnel to warm up...\n")
		time.Sleep(3 * time.Second)

		fmt.Printf("🔌 Connecting to DB on 127.0.0.1:%d...\n", localPort)
		fmt.Printf("\n--- DATABASE SESSION START ---\n")
		err = utils.RunPsql(dbUser, dbPass, dbName, "127.0.0.1", localPort)
		fmt.Printf("--- DATABASE SESSION END ---\n\n")
		
		if err != nil {
			fmt.Printf("Database session ended with error: %v\n", err)
		} else {
			fmt.Println("Database session closed successfully.")
		}
	},
}

var rdsCheckAccessCmd = &cobra.Command{
	Use:   "check-access [filter]",
	Short: "Check which instances have network access to an RDS",
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

		instances, err := aws.ListRDSInstances(cfg, filter)
		if err != nil {
			fmt.Printf("Error listing RDS: %v\n", err)
			return
		}

		var selectedRDS string
		if len(instances) == 1 {
			selectedRDS = instances[0]
		} else {
			selectedRDS, err = ui.SelectOne("Select RDS to check", instances)
			if err != nil || selectedRDS == "" {
				return
			}
		}

		fmt.Printf("Checking access rules for %s...\n", selectedRDS)
		info, err := aws.CheckRDSAccess(cfg, selectedRDS)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("\n" + titleStyle.Render("─── SECURITY GROUP ANALYSIS ───"))
		fmt.Printf("%s %s\n", keyStyle.Render("Allowed SGs:"), valStyle.Render(strings.Join(info.AllowedSGs, ", ")))
		fmt.Printf("%s %s\n", keyStyle.Render("Allowed IPs:"), valStyle.Render(strings.Join(info.AllowedIPs, ", ")))
		
		if len(info.Instances) == 0 {
			fmt.Println("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render("⚠️ No running EC2 instances found with direct SG access."))
		} else {
			fmt.Println("\n" + titleStyle.Render("─── INSTANCES WITH ACCESS ───"))
			for _, inst := range info.Instances {
				fmt.Printf("• %s (%s) - %s [SG: %s]\n", 
					valStyle.Render(inst.Name), 
					infoStyle.Render(inst.ID), 
					infoStyle.Render(inst.PrivateIP),
					infoStyle.Render(inst.SGID))
			}
		}
		fmt.Println("\n" + titleStyle.Render("──────────────────────────────"))
	},
}

func init() {
	rootCmd.AddCommand(rdsCmd)
	rdsCmd.AddCommand(rdsListCmd)
	rdsCmd.AddCommand(rdsConnCmd)
	rdsCmd.AddCommand(rdsCheckAccessCmd)
}

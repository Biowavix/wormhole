/*
Package cmd handles the CLI command definitions using Cobra.
root.go defines the base command and global flags like profile and region.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
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

	// Estilos de Lipgloss
	var (
		primaryColor = lipgloss.Color("#7D56F4") // Violeta Eléctrico
		descStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Italic(true)
		headerStyle  = lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	)

	// banner := `...` (comentado a petición del usuario)

	header := descStyle.Render("Secure infrastructure bridge for private cloud resource access via SSM tunneling.")

	rootCmd.SetHelpTemplate(fmt.Sprintf(`%s

%s
  {{.UseLine}}

%s
{{if .HasAvailableSubCommands}}{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}  %-12s {{.Short}}
{{end}}{{end}}{{end}}
%s
{{if .HasAvailableLocalFlags}}{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}

%s
  Use "{{.CommandPath}} [command] --help" for more information about a command.
`,
		header,
		headerStyle.Render("Usage:"),
		headerStyle.Render("Available Commands:"),
		"{{rpad .Name .NamePadding }}",
		headerStyle.Render("Flags:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render("Help:"),
	))
}

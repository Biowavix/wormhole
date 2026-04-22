/*
Package ui handles the visual representation of data in the terminal.
This file specifically handles tabular output using Lipgloss and Charm's components.
*/
package ui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)
	
	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			PaddingLeft(2)
)

// PrintList displays a simple list of items with a header.
func PrintList(header string, items []string) {
	fmt.Println(headerStyle.Render(header))
	for i, item := range items {
		fmt.Printf("%d. %s\n", i+1, itemStyle.Render(item))
	}
}

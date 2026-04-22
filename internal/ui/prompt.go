/*
Package ui handles terminal interactivity.
This file specifically handles user prompts using Huh.
*/
package ui

import (
	"github.com/charmbracelet/huh"
)

// SelectOne displays a list of options and returns the selected one.
func SelectOne(title string, options []string) (string, error) {
	if len(options) == 0 {
		return "", nil
	}
	
	var selected string
	var huhOptions []huh.Option[string]
	for _, opt := range options {
		huhOptions = append(huhOptions, huh.NewOption(opt, opt))
	}
	
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(huhOptions...).
				Value(&selected),
		),
	)
	
	err := form.Run()
	return selected, err
}

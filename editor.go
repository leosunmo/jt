package jt

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	DefaultEditor = "vim"

	boilerPlate = `%s%s
# Please enter the issue summary on the first line.
# Separate the summary from the description with an empty line.
# Lines starting with '#' will be ignored.
# The rest of the file will be used as the issue description.
`
)

var (
	ErrEmptySummary = fmt.Errorf("aborting, summary empty")
)

// OpenInEditor opens the user's default editor and returns the contents of the
// file.
func OpenInEditor(s string, d string) (string, string, error) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "*-ISSUE_MSG.jt")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temporary file: %s", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write the template to the file
	if d != "" {
		d = "\n\n" + d
	}
	_, err = tmpfile.WriteString(fmt.Sprintf(boilerPlate, s, d))
	if err != nil {
		return "", "", fmt.Errorf("failed to write boilerplate to file: %s", err)
	}
	// Open the file in EDITOR
	e := os.Getenv("EDITOR")
	if e == "" {
		e = DefaultEditor
	}
	cmd := exec.Command(e, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return "", "", err
	}

	// Read the resulting file
	result, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", "", err
	}

	// Parse the file into summary and description
	lines := strings.Split(string(result), "\n")
	var summary, description string
	for _, line := range lines {
		// Skip empty lines and comments
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		// Assign the first valid line to summary
		if summary == "" {
			summary = line
			continue
		}

		// Assign the rest to description
		description += line + "\n"
	}

	// Check if we have a summary
	if summary == "" {
		return "", "", ErrEmptySummary
	}

	return summary, description, nil
}

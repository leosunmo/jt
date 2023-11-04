package jt

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	DefaultEditor = "vim"

	boilerPlate = `
# Please enter the issue summary on the first line.
# Separate the summary from the description with an empty line.
# Lines starting with '#' will be ignored.
# The rest of the file will be used as the issue description.
`
)

var (
	ErrNotModified = fmt.Errorf("file not modified")
)

// OpenInEditor opens the user's default editor and returns the contents of the
// file.
func OpenInEditor() (string, string, error) {
	// create a temporary file
	tmpfile, err := os.CreateTemp("", "*-ISSUE_MSG.jt")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temporary file: %s", err)
	}
	defer os.Remove(tmpfile.Name())

	// write the template to the file
	_, err = tmpfile.WriteString(boilerPlate)
	if err != nil {
		return "", "", fmt.Errorf("failed to write boilerplate to file: %s", err)
	}
	// open the file in EDITOR
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

	// read the resulting file
	result, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", "", err
	}

	// Check if the file has been modified
	if string(result) == boilerPlate {
		return "", "", ErrNotModified
	}

	// split the file into summary and description
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
	return summary, description, nil
}

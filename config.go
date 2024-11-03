package jt

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigLocation = "~/.config/jt/config.yaml"
)

type JTConfig struct {
	// URL is the URL of the JIRA instance.
	URL string `yaml:"url"`
	// Email is the JIRA user email. Used as a username for authenticating.
	Email string `yaml:"email"`
	// Default project key is the JIRA project that will be used for issues. This is the short version of a project name, example: PRJ.
	DefaultProjectKey string `yaml:"defaultProjectKey"`
	// Default issue type is the issue type that will be used for issues.
	DefaultIssueType string `yaml:"defaultIssueType"`
	// Default component names are the default components that will be added to issues.
	DefaultComponentNames []string `yaml:"defaultComponentNames"`
	// Default parent issue types are the issue types that will be searched for when querying for parent issues.
	DefaultParentIssueTypes []string `yaml:"defaultParentIssueTypes"`
}

// ReadConfig reads config file from the default location.
func ReadConfig(configPath string) (JTConfig, error) {
	c := JTConfig{}

	f, err := os.Open(expandPath(configPath))
	if err != nil {
		return c, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&c); err != nil {
		return c, fmt.Errorf("failed to decode config file: %w", err)
	}

	return c, nil
}

func expandPath(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		return filepath.Join(dir, path[2:])
	}
	return path
}

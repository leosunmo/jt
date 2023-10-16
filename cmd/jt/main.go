package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/leosunmo/jt"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Read the issue summary from the command line arguments.
	summary := strings.Join(os.Args[1:], " ")
	if summary == "" {
		return fmt.Errorf("please provide a summary for the issue")
	}

	// Get the token from the keyring.
	t, err := jt.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get token: %s\n", err)
	}

	conf, err := jt.ReadConfig(jt.DefaultConfigLocation)
	if err != nil {
		return fmt.Errorf("failed to read config: %s\n", err)
	}

	parsedURL, err := url.Parse(conf.URL)
	if err != nil {
		return fmt.Errorf("failed to parse URL, %w", err)
	}

	jc := jt.JiraConfig{
		URL:            parsedURL.String(),
		Email:          conf.Email,
		Token:          t,
		ProjectKey:     conf.DefaultProjectKey,
		IssueType:      conf.DefaultIssueType,
		ComponentNames: conf.DefaultComponentNames,
	}

	c := jt.NewJiraClient(jc)
	key, err := c.NewJIRATicket(summary)
	if err != nil {
		return fmt.Errorf("failed to create ticket: %s\n", err)
	}

	fmt.Printf("created ticket: %s\tURL: %s\n", key, parsedURL.String()+"/browse/"+key)
	return nil
}

package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/leosunmo/jt"
)

var (
	msg = flag.String("m", "", "Issue Description, optional")
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse flags
	flag.Parse()

	var desc string
	// Check if msg is set
	if *msg != "" {
		desc = *msg
	}

	// Read the issue summary from the command line arguments.
	summary := strings.Join(flag.Args(), " ")
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
	key, err := c.NewJIRATicket(summary, desc)
	if err != nil {
		return fmt.Errorf("failed to create ticket: %s\n", err)
	}

	fmt.Printf("created ticket: %s\tURL: %s\n", key, parsedURL.String()+"/browse/"+key)
	return nil
}

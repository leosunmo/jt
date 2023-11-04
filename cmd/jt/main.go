package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/leosunmo/jt"
	"github.com/spf13/pflag"
)

var (
	msg  = pflag.StringP("msg", "m", "", "issue description, optional")
	edit = pflag.BoolP("edit", "e", false, "open the issue in your default editor, optional")
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	fl := pflag.NewFlagSet("jt", pflag.ContinueOnError)

	fl.Usage = func() {
		fmt.Println("Usage: jt [flags] [summary]")
		fmt.Println("\nIf summary is not provided, jt will open your default editor and prompt you for a summary and description.")
		fmt.Println("\nFlags:")
		pflag.PrintDefaults()
	}

	// Parse flags
	err := fl.Parse(os.Args[1:])
	if err != nil {
		if !errors.Is(err, pflag.ErrHelp) {
			fl.Usage()
			fmt.Printf("\n%s\n", err)
		}
		return nil
	}

	var desc string
	// Check if msg is set
	if *msg != "" {
		desc = *msg
	}

	// Read the issue summary from the command line arguments.
	summary := strings.Join(pflag.Args(), " ")

	if summary == "" || *edit {
		var err error
		summary, desc, err = jt.OpenInEditor(summary, desc)
		if err != nil {
			return err
		}
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

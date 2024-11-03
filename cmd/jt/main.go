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
	issueFlags     = pflag.NewFlagSet("issues", pflag.ContinueOnError)
	exclusiveFlags = pflag.NewFlagSet("exclusive", pflag.ContinueOnError)
	msg            = issueFlags.StringP("msg", "m", "", "Issue description, optional")
	edit           = issueFlags.BoolP("edit", "e", false, "Open default editor for summary and description, optional")
	parent         = issueFlags.StringP("parent", "p", "", "Assign the issue to a parent Epic or Initiative, optional")
	query          = exclusiveFlags.StringSliceP("query", "q", []string{}, `Query issues and exit. Available queries are: "parents", "epics", "initiatives", "tasks", and "bugs".
The "parents" query will search for parent issues (Epics, Initiatives by default).
A wildcard text search term can also be provided after a comma.
For example: jt -q "parents,some issue". Double quote if the search text contains spaces.`)
	completion = exclusiveFlags.BoolP("completion", "c", false, "Print zsh shell completion script to stdout and exit")
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	rootFlags := pflag.NewFlagSet("root", pflag.ContinueOnError)
	rootFlags.Usage = func() {
		fmt.Println("Usage: jt [flags] [summary]")
		fmt.Println("\nIf summary is not provided, jt will open your default editor and prompt you for a summary and description.")
		fmt.Println("\nIssue Creation Flags:")
		issueFlags.PrintDefaults()
		fmt.Println("\nExclusive Flags:")
		exclusiveFlags.PrintDefaults()
	}
	rootFlags.AddFlagSet(issueFlags)
	rootFlags.AddFlagSet(exclusiveFlags)
	// Parse flags
	err := rootFlags.Parse(os.Args[1:])
	if err != nil {
		if !errors.Is(err, pflag.ErrHelp) {
			rootFlags.Usage()
			fmt.Printf("\n%s\n", err)
		}
		return nil
	}

	if *completion {
		printCompletionZSH()
		return nil
	}

	// If a query is provided, run the query and return.
	if len(*query) > 0 {
		return runQuery(*query)
	}

	var desc string
	// Check if msg is set
	if *msg != "" {
		desc = *msg
	}

	// Read the issue summary from the command line arguments.
	summary := strings.Join(rootFlags.Args(), " ")

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
		URL:   parsedURL.String(),
		Email: conf.Email,
		Token: t,
	}

	ic := jt.IssueConfig{
		Summary:        summary,
		Description:    desc,
		ProjectKey:     conf.DefaultProjectKey,
		IssueType:      conf.DefaultIssueType,
		ComponentNames: conf.DefaultComponentNames,
	}

	if parent != nil && *parent != "" {
		ic.ParentIssueKey = *parent
	}

	c := jt.NewJiraClient(jc)
	key, err := c.NewJIRAIssue(ic)
	if err != nil {
		return fmt.Errorf("failed to create issue: %s\n", err)
	}

	fmt.Printf("created issue: %s\tURL: %s\n", key, parsedURL.String()+"/browse/"+key)
	return nil
}

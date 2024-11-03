package main

import (
	"fmt"
	"net/url"
	"sort"

	"github.com/leosunmo/jt"
	"github.com/leosunmo/jt/jql"
)

func runQuery(queryStrings []string) error {
	// Split the queryStrings to get the query type from the first element
	queryType := queryStrings[0]

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

	c := jt.NewJiraClient(jc)

	qb := jql.NewBuilder()
	qb.
		Equals("project", conf.DefaultProjectKey).
		And().
		In("component", conf.DefaultComponentNames...)

	// If a string is provided after the `,`, add it as a summary search term with a wildcard.
	if len(queryStrings) > 1 {
		qb.And().Contains("summary", queryStrings[1]+"*")
	}

	switch queryType {
	case "parents":
		// Query for parent issues (Epics, Initiatives by default)
		if conf.DefaultParentIssueTypes != nil {
			qb.And().In("type", conf.DefaultParentIssueTypes...)
		} else {
			qb.And().In("type", jt.IssueTypeEpic, jt.IssueTypeInitiative)
		}
		parents, err := doQuery(c, qb)
		if err != nil {
			return fmt.Errorf("failed to query parents: %s\n", err)
		}
		for _, parent := range parents {
			fmt.Printf("%s\n", parent)
		}
	case "epics":
		// Query for Epics and Initiatives
		qb.And().Equals("type", jt.IssueTypeEpic)
		e, err := doQuery(c, qb)
		if err != nil {
			return fmt.Errorf("failed to query epics: %s\n", err)
		}
		for _, epic := range e {
			fmt.Printf("%s\n", epic)
		}
	case "initiatives":
		// Query for Initiatives
		qb.And().Equals("type", jt.IssueTypeInitiative)
		initiatives, err := doQuery(c, qb)
		if err != nil {
			return fmt.Errorf("failed to query initiatives: %s\n", err)
		}
		for _, initiative := range initiatives {
			fmt.Printf("%s\n", initiative)
		}
	case "tasks":
		// Query for Tasks and Bugs
		qb.And().Equals("type", jt.IssueTypeTask)
		tasks, err := doQuery(c, qb)
		if err != nil {
			return fmt.Errorf("failed to query tasks: %s\n", err)
		}
		for _, task := range tasks {
			fmt.Printf("%s\n", task)
		}
	case "bugs":
		// Query for Bugs
		qb.And().Equals("type", jt.IssueTypeBug)
		bugs, err := doQuery(c, qb)
		if err != nil {
			return fmt.Errorf("failed to query bugs: %s\n", err)
		}
		for _, bug := range bugs {
			fmt.Printf("%s\n", bug)
		}
	default:
		return fmt.Errorf("unsupported query type: %s", queryType)
	}
	return nil
}

func doQuery(c *jt.JiraClient, qb *jql.JQLQueryBuilder) ([]string, error) {
	q, err := qb.Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build query: %s\n", err)
	}

	queryReq := jt.JQLSearchRequest{
		JQL: q,
		IncludedFields: []jt.Field{
			jt.FieldComponents,
			jt.FieldIssuetype,
			jt.FieldSummary,
		},
	}

	issues, err := c.SearchJiraIssues(queryReq)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues: %s\n", err)
	}
	// Sort issues by type: Initiatives first, then Epics, then Stories, lastly Tasks
	sortOrder := map[string]int{
		jt.IssueTypeInitiative: 1,
		jt.IssueTypeEpic:       2,
		jt.IssueTypeStory:      3,
		jt.IssueTypeTask:       4,
	}

	sort.SliceStable(issues, func(i, j int) bool {
		return sortOrder[issues[i].Fields.Issuetype.Name] < sortOrder[issues[j].Fields.Issuetype.Name]
	})

	output := make([]string, 0, len(issues))

	for _, issue := range issues {
		output = append(output, fmt.Sprintf("%s [%s]: %s", issue.Key, issue.Fields.Issuetype.Name, issue.Fields.Summary))
	}
	return output, nil
}

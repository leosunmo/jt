package jt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type JiraConfig struct {
	URL   string
	Email string
	Token string
}

type JiraClient struct {
	c      *http.Client
	config JiraConfig
}

type basicAuthTransport struct {
	username string
	password string
}

func (t *basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.username, t.password)
	return http.DefaultTransport.RoundTrip(req)
}

func NewJiraClient(conf JiraConfig) *JiraClient {
	return &JiraClient{
		c: &http.Client{
			Transport: &basicAuthTransport{
				username: conf.Email,
				password: conf.Token,
			},
		},
		config: conf,
	}
}

type IssueConfig struct {
	Summary        string
	Description    string
	ProjectKey     string
	IssueType      string
	ComponentNames []string
	ParentIssueKey string
}

type CreateIssueRequest struct {
	Fields Fields `json:"fields,omitempty"`
	Update struct {
		Labels []string `json:"labels,omitempty"`
	} `json:"update"`
}

const (
	IssueTypeBug        = "Bug"
	IssueTypeTask       = "Task"
	IssueTypeStory      = "Story"
	IssueTypeEpic       = "Epic"
	IssueTypeSubTask    = "Sub-task"
	IssueTypeInitiative = "Initiative"
)

type Field string

const (
	FieldSummary     Field = "summary"
	FieldDescription Field = "description"
	FieldProject     Field = "project"
	FieldIssuetype   Field = "issuetype"
	FieldComponents  Field = "components"
	FieldParent      Field = "parent"
)

type Fields struct {
	Components  []Components `json:"components,omitempty"`
	Issuetype   Issuetype    `json:"issuetype,omitempty"`
	Parent      *Parent      `json:"parent,omitempty"`
	Project     Project      `json:"project,omitempty"`
	Description *Description `json:"description,omitempty"`
	Summary     string       `json:"summary,omitempty"`
}

type Components struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
type Issuetype struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description"`
	Subtask     bool   `json:"subtask"`
}

type Parent struct {
	Key string `json:"key,omitempty"`
}

type Project struct {
	ID  string `json:"id,omitempty"`
	Key string `json:"key,omitempty"`
}
type ContentBlock struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}
type Attrs struct {
}
type Content struct {
	Type    string         `json:"type,omitempty"`
	Content []ContentBlock `json:"content,omitempty"`
}
type Description struct {
	Version int       `json:"version,omitempty"`
	Type    string    `json:"type,omitempty"`
	Content []Content `json:"content,omitempty"`
}

type JQLSearchRequest struct {
	JQL            string  `json:"jql"`
	IncludedFields []Field `json:"fields"`
	NextPageToken  string  `json:"nextPageToken,omitempty"`
}

type JQLSearchResponse struct {
	Issues        []Issue `json:"issues"`
	NextPageToken string  `json:"nextPageToken,omitempty"`
}

type Issue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Self   string `json:"self"`
	Fields Fields `json:"fields,omitempty"`
}

type Component struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreatedIssueResponse struct {
	ID         string `json:"id"`
	Key        string `json:"key"`
	Self       string `json:"self"`
	Transition struct {
		Status          int `json:"status"`
		ErrorCollection struct {
			ErrorMessages []string `json:"errorMessages"`
			Errors        struct{} `json:"errors"`
		} `json:"errorCollection"`
	} `json:"transition"`
	ErrorMessages []string          `json:"errorMessages"`
	Errors        map[string]string `json:"errors"`
}

// NewJIRAIssue creates a new JIRA issue using the JIRA REST API v3.
// The function returns the key of the created issue and an error if the issue could not be created.
// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/#api-rest-api-3-issue-post
func (jc JiraClient) NewJIRAIssue(conf IssueConfig) (string, error) {
	// Build the body of the request using a CreateIssueRequest
	reqBody := CreateIssueRequest{}

	reqBody.Fields.Summary = conf.Summary
	reqBody.Fields.Project.Key = conf.ProjectKey
	reqBody.Fields.Issuetype.Name = conf.IssueType

	reqBody.Fields.Components = make([]Components, len(conf.ComponentNames))
	for i, name := range conf.ComponentNames {
		reqBody.Fields.Components[i].Name = name
	}

	if conf.Description != "" {
		reqBody.Fields.Description = setDescription(conf.Description)
	}

	if conf.ParentIssueKey != "" {
		reqBody.Fields.Parent = &Parent{Key: conf.ParentIssueKey}
	}

	jsonBody, err := json.MarshalIndent(reqBody, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal body, %w", err)
	}

	req, err := http.NewRequest("POST", jc.config.URL+"/rest/api/3/issue", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request, %w", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := jc.c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body, %w", err)
	}

	createResponse := CreatedIssueResponse{}
	err = json.Unmarshal(b, &createResponse)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response, %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		var errMsg string
		if len(createResponse.ErrorMessages) != 0 {
			errMsg = strings.Join(createResponse.ErrorMessages, ", ")
		} else {
			for k, v := range createResponse.Errors {
				errMsg = errMsg + fmt.Sprintf("%s: %s", k, v)
			}
		}
		return "", fmt.Errorf("non-200 status %d, %s", resp.StatusCode, errMsg)
	}

	return createResponse.Key, nil
}

func setDescription(msg string) *Description {
	desc := Description{
		Type:    "doc",
		Version: 1,
		Content: []Content{
			{
				Type: "paragraph",
				Content: []ContentBlock{
					{
						Type: "text",
						Text: msg,
					},
				},
			},
		},
	}
	return &desc
}

// SearchJiraIssues searches for JIRA issues using the JIRA REST API v3.
// The function returns a slice of JQLSearchResponse and an error if the search request failed.
func (jc JiraClient) SearchJiraIssues(jqlReq JQLSearchRequest) ([]Issue, error) {
	var allIssues []Issue
	var nextPageToken string

	for {
		reqBody := struct {
			JQL           string   `json:"jql"`
			Fields        []string `json:"fields"`
			NextPageToken string   `json:"nextPageToken,omitempty"`
		}{
			JQL:           jqlReq.JQL,
			Fields:        convertFields(jqlReq.IncludedFields),
			NextPageToken: nextPageToken,
		}

		queryResp, err := jc.doJiraSearchRequest(reqBody)
		if err != nil {
			return nil, fmt.Errorf("search request failed: %w", err)
		}

		allIssues = append(allIssues, queryResp.Issues...)

		// Check for the next page token
		if queryResp.NextPageToken == "" {
			break // No more pages, exit the loop
		}
		nextPageToken = queryResp.NextPageToken
	}

	return allIssues, nil
}

// doJiraSearchRequest is a helper to perform the request and handle pagination token
func (jc JiraClient) doJiraSearchRequest(reqBody interface{}) (JQLSearchResponse, error) {
	var queryResp JQLSearchResponse

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return queryResp, fmt.Errorf("failed to marshal body, %w", err)
	}

	req, err := http.NewRequest("POST", jc.config.URL+"/rest/api/3/search/jql", bytes.NewReader(jsonBody))
	if err != nil {
		return queryResp, fmt.Errorf("failed to create request, %w", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := jc.c.Do(req)
	if err != nil {
		return queryResp, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return queryResp, fmt.Errorf("failed to read body, %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return queryResp, fmt.Errorf("non-200 status %d\nmessage: %s", resp.StatusCode, string(b))
	}

	if err := json.Unmarshal(b, &queryResp); err != nil {
		return queryResp, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return queryResp, nil
}

// convertFields converts the IncludedFields to a slice of strings
func convertFields(fields []Field) []string {
	result := make([]string, len(fields))
	for i, f := range fields {
		result[i] = string(f)
	}
	return result
}

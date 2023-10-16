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
	URL            string
	Email          string
	Token          string
	ProjectKey     string
	IssueType      string
	ComponentNames []string
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

type CreateIssueRequest struct {
	Fields Fields `json:"fields,omitempty"`
	Update struct {
		Labels []string `json:"labels,omitempty"`
	} `json:"update"`
}

type Fields struct {
	Components  []Components `json:"components,omitempty"`
	Issuetype   Issuetype    `json:"issuetype,omitempty"`
	Project     Project      `json:"project,omitempty"`
	Description *Description `json:"description,omitempty"`
	Summary     string       `json:"summary,omitempty"`
}

type Components struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
type Issuetype struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
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
	Type    string    `json:"type,omitempty"`
	Content []Content `json:"content,omitempty"`
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

// NewJIRATicket creates a new JIRA ticket using the JIRA REST API v3.
// The function returns the key of the created ticket and an error if the ticket could not be created.
// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/#api-rest-api-3-issue-post
func (jc JiraClient) NewJIRATicket(summary string) (string, error) {

	// Build the body of the request using a CreateIssueRequest
	reqBody := CreateIssueRequest{}

	reqBody.Fields.Project.Key = jc.config.ProjectKey
	reqBody.Fields.Issuetype.Name = jc.config.IssueType
	reqBody.Fields.Components = make([]Components, len(jc.config.ComponentNames))
	for i, name := range jc.config.ComponentNames {
		reqBody.Fields.Components[i].Name = name
	}

	reqBody.Fields.Summary = summary

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

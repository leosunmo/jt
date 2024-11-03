package jql

import (
	"strings"
	"testing"
)

func TestValidSimpleQuery(t *testing.T) {

	qb := NewBuilder()

	// Construct the following query:
	// project = 'My Project' AND status IN ('In Progress','To do') ORDER BY created DESC
	qb.Equals("project", "My Project").And().In("status", "In Progress", "To do").OrderBy(false, "created")

	s, err := qb.Build()
	if err != nil {
		t.Fatalf("failed to build query: %s", err)
	}

	expected := `project = 'My Project' AND status IN ('In Progress', 'To do') ORDER BY created DESC`
	if s != expected {
		t.Fatalf("expected %q, got %q", expected, s)
	}
}

func TestValidComplexQuery(t *testing.T) {
	builder := NewBuilder()

	query, err := builder.
		// Add conditions with Equals, NotEquals, In, and NotIn
		Equals("status", "Open").
		And().
		NotEquals("priority", "Low").
		And().
		In("assignee", "Alice", "Bob", "Charlie").
		And().
		NotIn("project", []string{"ProjectA", "ProjectB"}).

		// Use OR to add branching conditions
		Or().
		Equals("reporter", "Dave").
		And().
		NotEquals("issueType", "Bug").

		// Nested conditions with different operators
		Or().
		In("labels", "critical", "urgent").
		And().
		Equals("resolution", "Unresolved").

		// More conditions to increase complexity
		And().
		Equals("created", "2023-10-01").
		And().
		NotEquals("updated", "2023-11-01").
		And().
		In("component", "Backend", "Frontend").

		// Add ORDER BY at the end
		OrderBy(true, "priority", "created").
		Build()

	if err != nil {
		t.Fatalf("failed to build query: %s", err)
	}

	expQuery := `status = 'Open' AND priority != 'Low' AND assignee IN ('Alice', 'Bob', 'Charlie') AND project NOT IN ('ProjectA', 'ProjectB') OR reporter = 'Dave' AND issueType != 'Bug' OR labels IN ('critical', 'urgent') AND resolution = 'Unresolved' AND created = '2023-10-01' AND updated != '2023-11-01' AND component IN ('Backend', 'Frontend') ORDER BY priority, created ASC`

	if query != expQuery {
		t.Fatalf("expected %q, got %q", expQuery, query)
	}
}
func TestInvalidQueries(t *testing.T) {
	testData := []struct {
		name   string
		fn     func(*JQLQueryBuilder) *JQLQueryBuilder
		errMsg string
	}{
		{
			name: "consecutive operators",
			fn: func(q *JQLQueryBuilder) *JQLQueryBuilder {
				return q.Equals("status", "Open").Equals("priority", "High")
			},
			errMsg: "consecutive operators",
		},
		{
			// Invalid because "ORDER BY" must follow an operator and not a keyword like "AND"
			// Query: "status = 'Open' AND ORDER BY created ASC"
			name: "ORDER BY after keyword",
			fn: func(q *JQLQueryBuilder) *JQLQueryBuilder {
				return q.Equals("status", "Open").And().OrderBy(false, "created")
			},
			errMsg: "consecutive keywords",
		},
		{
			// Invalid because "ORDER BY" should not follow another "ORDER BY" directly
			// Query: "ORDER BY created ASC ORDER BY updated DESC"
			name: "consecutive ORDER BY",
			fn: func(q *JQLQueryBuilder) *JQLQueryBuilder {
				return q.Equals("status", "Open").OrderBy(true, "created").OrderBy(false, "updated")
			},
			errMsg: "consecutive keywords",
		},
		{
			// Invalid because an operator cannot directly follow a keyword like "OR"
			// Query: "OR = 'Open'"
			name: "operator after keyword",
			fn: func(q *JQLQueryBuilder) *JQLQueryBuilder {
				return q.Or().Equals("status", "Open")
			},
			errMsg: "first word must be an operator",
		},
		{
			// Invalid because two operators cannot appear consecutively
			// Query: "status = 'Open' IN 'In Progress'"
			name: "consecutive operators",
			fn: func(q *JQLQueryBuilder) *JQLQueryBuilder {
				return q.Equals("status", "Open").In("priority")
			},
			errMsg: "consecutive operators",
		},
		{
			// Invalid because the "ORDER BY" keyword cannot be the first part of the query
			// Query: "ORDER BY created ASC status = 'Open'"
			name: "ORDER BY at the start",
			fn: func(q *JQLQueryBuilder) *JQLQueryBuilder {
				return q.OrderBy(false, "created").Equals("status", "Open")
			},
			errMsg: "first word must be an operator",
		},
		{
			// Invalid because "ORDER BY" must be the last part of the query
			// Query: "status = 'Open' ORDER BY created ASC AND priority = 'High'"
			name: "ORDER BY in the middle",
			fn: func(q *JQLQueryBuilder) *JQLQueryBuilder {
				return q.Equals("status", "Open").OrderBy(false, "created").And().Equals("priority", "High")
			},
			errMsg: "must be the last part of the query",
		},
		{
			// Invalid because a query cannot contain only "ORDER BY" without a preceding opertor
			// Query: "ORDER BY created ASC"
			name: "only ORDER BY",
			fn: func(q *JQLQueryBuilder) *JQLQueryBuilder {
				return q.OrderBy(true, "created")
			},
			errMsg: "first word must be an operator",
		},
		{
			// Invalid because a query cannot end with a keyword
			// Query: "status = 'Open' AND priority = 'High' AND"
			name: "end with AND",
			fn: func(q *JQLQueryBuilder) *JQLQueryBuilder {
				return q.Equals("status", "Open").And().Equals("priority", "High").And()
			},
			errMsg: "query cannot end with a keyword",
		},
		{
			// Invalid because a query cannot start with a keyword
			// Query: "OR status = 'Open'"
			name: "start with OR",
			fn: func(q *JQLQueryBuilder) *JQLQueryBuilder {
				return q.Or().Equals("status", "Open")
			},
			errMsg: "first word must be an operator",
		},
		{
			// Invalid because a query cannot contain consecutive keywords
			// Query: "status = 'Open' AND AND priority = 'High'"
			name: "consecutive keywords",
			fn: func(q *JQLQueryBuilder) *JQLQueryBuilder {
				return q.Equals("status", "Open").And().And().Equals("priority", "High")
			},
		},
		{
			// No query parts
			name:   "empty query",
			fn:     func(q *JQLQueryBuilder) *JQLQueryBuilder { return q },
			errMsg: "no query parts added",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewBuilder()
			_, err := tt.fn(qb).Build()
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Fatalf("expected error message to contain %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

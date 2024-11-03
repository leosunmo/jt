package jql

import (
	"fmt"
	"strings"
)

type jqlWord interface {
	Type() wordType
	String() string
}

type wordType int

const (
	unknownType wordType = iota
	operatorType
	keywordType
	endKeywordType
)

// Operator represents a basic field-operator-value component
type Operator struct {
	field    string
	operator string
	value    string
}

func (c Operator) String() string {
	return fmt.Sprintf("%s %s %s", c.field, c.operator, c.value)
}

func (c Operator) Type() wordType {
	return operatorType
}

// Keyword represents logical keywords such as AND, OR, etc.
type Keyword struct {
	name     string
	wordType wordType
	isEnding bool
}

func (k Keyword) String() string {
	return k.name
}

func (k Keyword) Type() wordType {
	return k.wordType
}

var (
	And            = Keyword{"AND", keywordType, false}
	Or             = Keyword{"OR", keywordType, false}
	Not            = Keyword{"NOT", keywordType, false}
	OrderByKeyword = Keyword{"ORDER BY", endKeywordType, true}
)

// OrderBy represents the ORDER BY component in JQL, with fields and sorting direction
type OrderBy struct {
	fields    []string
	ascending bool
}

func (o OrderBy) String() string {
	order := "ASC"
	if !o.ascending {
		order = "DESC"
	}
	return fmt.Sprintf("ORDER BY %s %s", strings.Join(o.fields, ", "), order)
}

func (o OrderBy) Type() wordType {
	return endKeywordType
}

// JQLQueryBuilder is the main query builder struct
type JQLQueryBuilder struct {
	qt             []jqlWord
	rawQueryString string
}

// NewJQLQuery initializes a new JQLQuery
func NewBuilder() *JQLQueryBuilder {
	return &JQLQueryBuilder{
		qt: make([]jqlWord, 0),
	}
}

func (q *JQLQueryBuilder) SetJQLString(jqlString string) *JQLQueryBuilder {
	q.rawQueryString = jqlString
	return q
}

func (q *JQLQueryBuilder) And() *JQLQueryBuilder {
	q.qt = append(q.qt, And)
	return q
}

func (q *JQLQueryBuilder) Or() *JQLQueryBuilder {
	q.qt = append(q.qt, Or)
	return q
}

// OrderBy adds an ORDER BY clause with specified fields and sorting direction
func (q *JQLQueryBuilder) OrderBy(ascending bool, fields ...string) *JQLQueryBuilder {
	q.qt = append(q.qt, OrderBy{fields: fields, ascending: ascending})
	return q
}

// Equals adds an equality operator for a field
func (q *JQLQueryBuilder) Equals(field string, value string) *JQLQueryBuilder {
	q.qt = append(q.qt, Operator{field: field, operator: "=", value: fmt.Sprintf("'%s'", value)})
	return q
}

// NotEquals adds an inequality operator for a field
func (q *JQLQueryBuilder) NotEquals(field string, value string) *JQLQueryBuilder {
	q.qt = append(q.qt, Operator{field: field, operator: "!=", value: fmt.Sprintf("'%s'", value)})
	return q
}

// In adds an IN operator for a field with a list of values
//
// field IN ('value1', 'value2', ...)
func (q *JQLQueryBuilder) In(field string, values ...string) *JQLQueryBuilder {
	valueList := "('" + strings.Join(values, "', '") + "')"
	q.qt = append(q.qt, Operator{field: field, operator: "IN", value: valueList})
	return q
}

// NotIn adds a NOT IN operator for a field with a list of values
func (q *JQLQueryBuilder) NotIn(field string, values []string) *JQLQueryBuilder {
	valueList := "('" + strings.Join(values, "', '") + "')"
	q.qt = append(q.qt, Operator{field: field, operator: "NOT IN", value: valueList})
	return q
}

func (q *JQLQueryBuilder) Contains(field string, value string) *JQLQueryBuilder {
	q.qt = append(q.qt, Operator{field: field, operator: "~", value: fmt.Sprintf("'%s'", value)})
	return q
}

// Build constructs the final JQL query string from the query parts
// It returns an error if the query is invalid.
// Only syntactic validation is done, stopping at the first detected error while still building the full query.
func (q *JQLQueryBuilder) Build() (string, error) {
	if len(q.qt) == 0 && q.rawQueryString == "" {
		return "", fmt.Errorf("no query parts added")
	}

	// If a raw query string was set, return it directly
	if q.rawQueryString != "" {
		return q.rawQueryString, nil
	}

	var builder strings.Builder
	var lastWord jqlWord
	var err error
	var errCharPos int
	currentPosition := 0

	for i, word := range q.qt {
		wordStr := word.String()

		if err == nil {
			// Validate the current word if we haven't encountered an error yet
			// If we have, we skip validation to avoid duplicate error messages
			err = validateWord(i, word, lastWord)
			errCharPos = currentPosition
		}

		// Append word to builder and update the current character position
		if builder.Len() > 0 {
			builder.WriteString(" ")
			currentPosition++ // Account for added space between words
		}
		builder.WriteString(wordStr)
		currentPosition += len(wordStr) // Account for word length
		lastWord = word
	}

	finalQuery := builder.String()

	// If we haven't encountered an error yet, validate the final query
	if err == nil {
		// Final validation: Ensure the query does not end with a keyword
		// Example of invalid query: "status = 'Open' AND"
		if lastWord.Type() == keywordType {
			err = fmt.Errorf("query cannot end with a keyword")
			errCharPos = currentPosition
		}
	}

	// If error was captured, return it with the full query and a pointer to the error position
	if err != nil {
		pointerLine := strings.Repeat(" ", errCharPos) + "^"
		return "", fmt.Errorf("invalid query:\n%q\n%s\nError: %s", finalQuery, pointerLine, err.Error())
	}

	return finalQuery, nil
}

func validateWord(i int, word jqlWord, lastWord jqlWord) error {
	// Validate the first word
	// The first word of the query must be an operator (e.g., "status = 'Open'").
	// If the first word is an operator, keyword, or end keyword, it's invalid.
	if i == 0 && word.Type() != operatorType {
		return fmt.Errorf("first word must be an operator, got %q", word.String())
	}

	if lastWord == nil {
		lastWord = &Keyword{"", unknownType, false}
	}

	// Validation rules based on last and current types
	switch word.Type() {
	case operatorType:
		// Check for consecutive operators
		// Two operators cannot appear consecutively without a keyword in between.
		// Example of invalid query: "status = 'Open' assignee = 'JohnDoe'"
		if lastWord.Type() == operatorType {
			return fmt.Errorf("consecutive operators %q & %q", lastWord.String(), word.String())
		}
	case keywordType:
		if lastWord.Type() == keywordType {
			return fmt.Errorf("consecutive keywords %q & %q", lastWord.String(), word.String())
		}
	case endKeywordType:
		// Ensure end keywords appear at the end of the query
		// End keywords (e.g., "ORDER BY") should only appear at the end.
		// Example of invalid query: "ORDER BY created ASC status = 'Open'"
		if lastWord.Type() != operatorType {
			return fmt.Errorf("consecutive keywords %q & %q", lastWord.String(), word.String())
		}
	}

	// Check if the lastWord was an end keyword
	if lastWord.Type() == endKeywordType {
		return fmt.Errorf("keyword %q must be the last part of the query", lastWord.String())
	}

	return nil
}

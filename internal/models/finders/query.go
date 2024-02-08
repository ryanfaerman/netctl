package finders

import (
	"fmt"
	"slices"
	"strings"
)

type QueryType int

const (
	QueryWhere QueryType = iota
	QueryField
)

// Query represents a query to be used by a Finder.
type Query struct {
	Fields []string
	Values []any
	Type   QueryType
}

func (q Query) String() string {
	return fmt.Sprintf("Query{Fields: %v, Values: %v, Type: %v}", q.Fields, q.Values, q.Type)
}

// QuerySet is a collection of queries
type QuerySet []Query

func (qs QuerySet) String() string {
	var b strings.Builder
	for _, q := range qs {
		b.WriteString(q.String())
		b.WriteString(";")
	}
	return b.String()
}

type QueryFunc func() (Query, error)

// Fields returns a list of all fields in the QuerySet
func (qs QuerySet) Fields() []string {
	var fields []string
	for _, q := range qs {
		fields = append(fields, q.Fields...)
	}
	return fields
}

// Of Type filters the queryset only to queries of a given type
func (qs QuerySet) OfType(t QueryType) QuerySet {
	var out QuerySet
	for _, q := range qs {
		if q.Type == t {
			out = append(out, q)
		}
	}

	return out
}

// HasField returns true if the QuerySet contains all of the given fields
func (qs QuerySet) HasField(fields ...string) bool {
	if len(fields) == 0 {
		return true
	}
	var has []bool
	for _, field := range fields {
		has = append(has, slices.Contains(qs.OfType(QueryField).Fields(), field))
	}
	return !slices.Contains(has, false)
}

// HasWhere returns true if the QuerySet contains all of the given fields for Where type queries
func (qs QuerySet) HasWhere(wheres ...string) bool {
	if len(wheres) == 0 {
		return true
	}

	var has []bool
	for _, where := range wheres {
		has = append(has, slices.Contains(qs.OfType(QueryWhere).Fields(), where))
	}
	return !slices.Contains(has, false)
}

// ValuesForField collects all values for a given field
func (qs QuerySet) ValuesForField(f string) []any {
	var out []any
	for _, q := range qs {
		if slices.Contains(q.Fields, f) {
			out = append(out, q.Values...)
		}
	}
	return out
}

// ValueForField returns the first value for a given field
func (qs QuerySet) ValueForField(f string) any {
	vals := qs.ValuesForField(f)
	if len(vals) > 0 {
		return vals[0]
	}
	return nil
}

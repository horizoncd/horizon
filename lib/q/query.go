package q

import "g.hz.netease.com/horizon/core/common"

// KeyWords ...
type KeyWords = map[string]interface{}

// Query parameters
type Query struct {
	// Filter list
	Keywords KeyWords
	// Sort list
	Sorts []*Sort
	// Page number
	PageNumber int
	// Page size
	PageSize int
}

func NewQuery(PageNumber, PageSize int, keywords KeyWords, sorts []*Sort) *Query {
	q := &Query{}
	if PageNumber < 1 {
		PageNumber = common.DefaultPageNumber
	}
	if PageSize < 1 {
		PageSize = common.DefaultPageSize
	}

	q.PageNumber = PageNumber
	q.PageSize = PageSize
	q.Keywords = keywords
	q.Sorts = sorts
	return q
}

// First make the query only fetch the first one record in the sorting order
func (q *Query) First(sorting ...*Sort) *Query {
	q.PageNumber = 1
	q.PageSize = 1
	if len(sorting) > 0 {
		q.Sorts = append(q.Sorts, sorting...)
	}

	return q
}

// New returns Query with keywords
func New(kw KeyWords) *Query {
	return &Query{Keywords: kw}
}

// MustClone returns the clone of query when it's not nil
// or returns a new Query instance
func MustClone(query *Query) *Query {
	q := &Query{
		Keywords: map[string]interface{}{},
	}
	if query != nil {
		q.PageNumber = query.PageNumber
		q.PageSize = query.PageSize
		q.Sorts = query.Sorts
		for k, v := range query.Keywords {
			q.Keywords[k] = v
		}
		for _, sort := range query.Sorts {
			q.Sorts = append(q.Sorts, &Sort{
				Key:  sort.Key,
				DESC: sort.DESC,
			})
		}
	}
	return q
}

// Sort specifies the order information
type Sort struct {
	Key  string
	DESC bool
}

// Range query
type Range struct {
	Min interface{}
	Max interface{}
}

// AndList query
type AndList struct {
	Values []interface{}
}

// OrList query
type OrList struct {
	Values []interface{}
}

// FuzzyMatchValue query
type FuzzyMatchValue struct {
	Value string
}

// NewSort creates new sort
func NewSort(key string, desc bool) *Sort {
	return &Sort{
		Key:  key,
		DESC: desc,
	}
}

// NewRange creates a new range
func NewRange(min, max interface{}) *Range {
	return &Range{
		Min: min,
		Max: max,
	}
}

// NewAndList creates a new and list
func NewAndList(values []interface{}) *AndList {
	return &AndList{
		Values: values,
	}
}

// NewOrList creates a new or list
func NewOrList(values []interface{}) *OrList {
	return &OrList{
		Values: values,
	}
}

// NewFuzzyMatchValue creates a new fuzzy match
func NewFuzzyMatchValue(value string) *FuzzyMatchValue {
	return &FuzzyMatchValue{
		Value: value,
	}
}

package database

import "context"

type Operator int

const (
	EQ Operator = iota
	NEQ
	LT
	LTE
	GT
	GTE
	IN
	NULL
	NOTNULL

	OR
	AND
)

type MatchItem struct {
	Field    string
	Operator Operator
	Value    interface{}
}

type Fields map[string]interface{}

type MatchOptions struct {
	Matches []MatchItem
	Sort    []string
	Limit   *int
	Offset  *int
}

type MatchOption func(*MatchOptions)

func (opts *MatchOptions) oper(field string, oper Operator, val interface{}) *MatchOptions {
	opts.Matches = append(opts.Matches, MatchItem{
		Field: field, Operator: oper, Value: val,
	})
	return opts
}

func (opts *MatchOptions) Apply(newOptions ...MatchOption) MatchOptions {
	for _, n := range newOptions {
		n(opts)
	}
	return *opts
}

func (opts *MatchOptions) OR(sub MatchOptions) *MatchOptions {
	opts.Matches = append(opts.Matches, MatchItem{Operator: OR, Value: sub})
	return opts
}

func (opts *MatchOptions) AND(sub MatchOptions) *MatchOptions {
	opts.Matches = append(opts.Matches, MatchItem{Operator: AND, Value: sub})
	return opts
}

func (opts *MatchOptions) EQ(field string, val interface{}) *MatchOptions {
	return opts.oper(field, EQ, val)
}

func (opts *MatchOptions) NEQ(field string, val interface{}) *MatchOptions {
	return opts.oper(field, NEQ, val)
}

func (opts *MatchOptions) LT(field string, val interface{}) *MatchOptions {
	return opts.oper(field, LT, val)
}

func (opts *MatchOptions) LTE(field string, val interface{}) *MatchOptions {
	return opts.oper(field, LTE, val)
}

func (opts *MatchOptions) GT(field string, val interface{}) *MatchOptions {
	return opts.oper(field, GT, val)
}

func (opts *MatchOptions) GTE(field string, val interface{}) *MatchOptions {
	return opts.oper(field, GTE, val)
}

func (opts *MatchOptions) Null(field string) *MatchOptions {
	return opts.oper(field, NULL, nil)
}

func (opts *MatchOptions) NotNull(field string) *MatchOptions {
	return opts.oper(field, NOTNULL, nil)
}

func (opts *MatchOptions) SetLimit(limit int) *MatchOptions {
	opts.Limit = &limit
	return opts
}

func (opts *MatchOptions) SetOffset(offset int) *MatchOptions {
	opts.Offset = &offset
	return opts
}

func (opts *MatchOptions) SetSort(sort ...string) *MatchOptions {
	opts.Sort = sort
	return opts
}

type RepositoryCodeGenerator interface {
	GenerateRepositoryCode(modelName, packageName string)
}

type Repository interface {
	// First get the first record of the records which fetched from the DB alongside the match condition
	First(ctx context.Context, v interface{}, opts ...MatchOption) error
	// Find record following the match condition
	Find(ctx context.Context, v interface{}, opts ...MatchOption) error
	// Count record following the match condition
	Count(ctx context.Context, v interface{}, count *int64, opts ...MatchOption) error
	// Update a record
	Update(ctx context.Context, v interface{}) error
	// Delete record following the match condition
	Delete(ctx context.Context, v interface{}, opts ...MatchOption) error
	// Create records
	Create(ctx context.Context, v interface{}) error
	// UpdateField field
	UpdateFields(ctx context.Context, v interface{}, fields Fields, opts ...MatchOption) error
	// Transaction
	Transaction(func() error)
}

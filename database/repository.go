package database

import (
	"context"
	"errors"
)

type Operator int

var (
	ErrRecordNotFound = errors.New("record not found")
)

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

	LeftJoin  = 0
	InnerJoin = 1
	RightJoin = 2
)

type MatchItem struct {
	Field    string
	Operator Operator
	Value    interface{}
}

type Field string

type Fields map[string]interface{}

type Join struct {
	Model interface{}
	Opts  MatchOptions
	Type  int
}

type Group struct {
	By     string
	Having *MatchOptions
}

type Model struct {
	Result interface{}
	From   interface{}
	Joins  []Join
	Grp    *Group
}

func M(result interface{}, froms ...interface{}) *Model {
	from := result
	if len(froms) > 0 {
		from = froms[0]
	}
	return &Model{Result: result, From: from}
}

func (m *Model) Group(group string, having ...MatchOption) *Model {
	m.Grp = &Group{By: group}
	if len(having) > 0 {
		m.Grp.Having = &MatchOptions{}
		for _, apply := range having {
			apply(m.Grp.Having)
		}
	}
	return m
}

func (m *Model) With(model interface{}, opts ...MatchOption) *Model {
	return m.with(model, LeftJoin, opts...)
}

func (m *Model) RWith(model interface{}, opts ...MatchOption) *Model {
	return m.with(model, RightJoin, opts...)
}

func (m *Model) IWith(model interface{}, opts ...MatchOption) *Model {
	return m.with(model, InnerJoin, opts...)
}

func (m *Model) with(model interface{}, j int, opts ...MatchOption) *Model {
	mj := Join{
		Model: model,
		Type:  j,
	}
	for _, apply := range opts {
		apply(&mj.Opts)
	}
	m.Joins = append(m.Joins, mj)
	return m
}

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

func (opts *MatchOptions) IN(field string, val interface{}) *MatchOptions {
	return opts.oper(field, IN, val)
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

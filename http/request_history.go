package http

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Store interface {
	Push(ctx context.Context, ip string, req Req) error
	Get(ctx context.Context, ip string, reqs *[]Req) error
}

type reqNode struct {
	next *reqNode
	data Req
}

type reqChain struct {
	head   *reqNode
	tail   *reqNode
	length int
}

type memoryStore struct {
	history map[string]*reqChain
	lock    sync.Mutex
	length  int
}

func NewMemoryStore(size int) Store {
	return &memoryStore{history: make(map[string]*reqChain), length: size}
}

func (m *memoryStore) Push(ctx context.Context, ip string, req Req) error {
	if chain, ok := m.history[ip]; ok {
		if chain.length >= m.length {
			head := chain.head
			chain.head = chain.head.next
			head.next = nil
			chain.length--
		}
		chain.tail.next = &reqNode{data: req}
		chain.tail = chain.tail.next
		chain.length++
		return nil
	}
	node := &reqNode{data: req}
	m.history[ip] = &reqChain{
		head:   node,
		tail:   node,
		length: 1,
	}
	return nil
}

func (m *memoryStore) Get(ctx context.Context, ip string, reqs *[]Req) error {
	if chain, ok := m.history[ip]; ok {
		*reqs = make([]Req, chain.length)
		node := chain.head
		i := 0
		for node != nil {
			(*reqs)[i] = node.data
			i++
		}
	}
	return nil
}

type requestHistory struct {
	store Store
}

func (his *requestHistory) Record(ctx context.Context, req *http.Request) {
	his.store.Push(ctx, req.RemoteAddr, Req{
		Method: req.Method,
		Query:  req.URL.Query(),
		Hash:   req.URL.Fragment,
	})
}

type Req struct {
	Method string
	Path   string
	Query  url.Values
	Hash   string
	Time   time.Time
}

type TailOptions Req

type TailOption func(*TailOptions)

func Path(path string) TailOption {
	return func(opts *TailOptions) {
		opts.Path = path
	}
}

func Query(query url.Values) TailOption {
	return func(opts *TailOptions) {
		opts.Query = query
	}
}

func Hash(hash string) TailOption {
	return func(opts *TailOptions) {
		opts.Hash = hash
	}
}

func Method(m string) TailOption {
	return func(opts *TailOptions) {
		opts.Method = m
	}
}

func (his *requestHistory) Tail(ctx context.Context, ip string, reqs *[]Req, opts ...TailOption) error {
	has := []Req{}
	opt := &TailOptions{}
	for _, apply := range opts {
		apply(opt)
	}
	if err := his.store.Get(ctx, ip, &has); err != nil {
		return err
	}
	for _, req := range has {
		if opt.Path != "" && req.Path != opt.Path {
			continue
		}
		if opt.Method != "" && req.Method != opt.Method {
			continue
		}
		*reqs = append(*reqs, req)
	}
	return nil
}

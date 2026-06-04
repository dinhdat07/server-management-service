package domain

import "context"

type ServerSearchFilter struct {
	Page          int
	Limit         int
	FilterStatus  string
	FilterName    string
	SortBy        string
	SortDirection string
}

type ServerSearchResult struct {
	Servers    []*Server
	TotalCount int64
}

type ServerSearchRepository interface {
	Search(ctx context.Context, filter ServerSearchFilter) (*ServerSearchResult, error)
}

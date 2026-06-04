package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	esv8 "github.com/elastic/go-elasticsearch/v8"
	"server-management-service/internal/modules/server_management/domain"
)

type ServerSearcher struct {
	client *esv8.Client
	index  string
}

func NewServerSearcher(client *esv8.Client, index string) *ServerSearcher {
	return &ServerSearcher{
		client: client,
		index:  index,
	}
}

func (s *ServerSearcher) Search(ctx context.Context, filter domain.ServerSearchFilter) (*domain.ServerSearchResult, error) {
	query := buildSearchQuery(filter)
	
	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("marshal search query: %w", err)
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(s.index),
		s.client.Search.WithBody(bytes.NewReader(body)),
		s.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	var esRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source ServerDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esRes); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	servers := make([]*domain.Server, 0, len(esRes.Hits.Hits))
	for _, hit := range esRes.Hits.Hits {
		doc := hit.Source
		servers = append(servers, &domain.Server{
			ServerID:            doc.ID,
			ServerName:          doc.ServerName,
			IPv4:                doc.IPv4,
			CurrentStatus:       domain.ServerStatus(doc.CurrentStatus),
			ConsecutiveFailures: doc.ConsecutiveFailures,
			CreatedAt:           doc.CreatedAt,
			UpdatedAt:           doc.UpdatedAt,
		})
	}

	return &domain.ServerSearchResult{
		Servers:    servers,
		TotalCount: esRes.Hits.Total.Value,
	}, nil
}

func buildSearchQuery(filter domain.ServerSearchFilter) map[string]any {
	must := []map[string]any{}

	if filter.FilterStatus != "" {
		must = append(must, map[string]any{
			"term": map[string]any{
				"current_status": filter.FilterStatus,
			},
		})
	}

	if filter.FilterName != "" {
		// Fuzzy search for server_name, and wildcard for ipv4
		must = append(must, map[string]any{
			"bool": map[string]any{
				"should": []map[string]any{
					{
						"match": map[string]any{
							"server_name": map[string]any{
								"query":     filter.FilterName,
								"fuzziness": "AUTO",
							},
						},
					},
					{
						"wildcard": map[string]any{
							"ipv4": fmt.Sprintf("*%s*", filter.FilterName),
						},
					},
				},
				"minimum_should_match": 1,
			},
		})
	}

	queryBody := map[string]any{}
	
	if len(must) > 0 {
		queryBody["query"] = map[string]any{
			"bool": map[string]any{
				"must": must,
			},
		}
	} else {
		queryBody["query"] = map[string]any{
			"match_all": map[string]any{},
		}
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	}

	queryBody["from"] = (filter.Page - 1) * filter.Limit
	queryBody["size"] = filter.Limit

	// Sorting
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	
	// Map frontend sort keys to ES index fields
	if sortBy == "name" || sortBy == "server_name" {
		sortBy = "server_name.keyword" // Need to sort on keyword field for text
	}

	sortDir := "desc"
	if filter.SortDirection == "asc" || filter.SortDirection == "ASC" {
		sortDir = "asc"
	}

	queryBody["sort"] = []map[string]any{
		{
			sortBy: map[string]any{
				"order": sortDir,
			},
		},
	}

	return queryBody
}

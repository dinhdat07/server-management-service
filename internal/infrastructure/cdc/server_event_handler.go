package cdc

import (
	"context"
	"encoding/json"
	"fmt"

	esinfra "server-management-service/internal/infrastructure/elasticsearch"
	redisinfra "server-management-service/internal/infrastructure/redis"
)

type ServerEventHandler struct {
	indexer *esinfra.ServerIndexer
	cache   *redisinfra.ServerCache
}

func NewServerEventHandler(indexer *esinfra.ServerIndexer, cache *redisinfra.ServerCache) *ServerEventHandler {
	return &ServerEventHandler{
		indexer: indexer,
		cache:   cache,
	}
}

func (h *ServerEventHandler) Handle(ctx context.Context, value []byte) error {
	var event DebeziumEvent
	if err := json.Unmarshal(value, &event); err != nil {
		return fmt.Errorf("unmarshal debezium event: %w", err)
	}

	switch event.Op {
	case "r", "c", "u":
		return h.handleUpsert(ctx, event)

	case "d":
		return h.handleDelete(ctx, event)

	default:
		return nil
	}
}

func (h *ServerEventHandler) handleUpsert(ctx context.Context, event DebeziumEvent) error {
	if event.After == nil {
		return nil
	}

	doc, err := esinfra.ServerDocumentFromDebeziumAfter(event.After)
	if err != nil {
		return fmt.Errorf("map server document: %w", err)
	}

	if doc.ID == "" {
		return nil
	}

	// 1. Sync to Elasticsearch
	if err := h.indexer.Index(ctx, doc); err != nil {
		return fmt.Errorf("index server document: %w", err)
	}

	// 2. Sync to Redis Cache
	if h.cache != nil {
		if err := h.cache.Upsert(ctx, doc.ID, doc.IPv4, doc.CurrentStatus, doc.ConsecutiveFailures); err != nil {
			return fmt.Errorf("upsert server cache: %w", err)
		}
	}

	return nil
}

func (h *ServerEventHandler) handleDelete(ctx context.Context, event DebeziumEvent) error {
	if event.Before == nil {
		return nil
	}

	id, ok := event.Before["server_id"].(string)
	if !ok || id == "" {
		return nil
	}

	// Delete from Elasticsearch
	if err := h.indexer.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete server document: %w", err)
	}

	// Delete from Redis Cache
	if h.cache != nil {
		if err := h.cache.Delete(ctx, id); err != nil {
			return fmt.Errorf("delete server cache: %w", err)
		}
	}

	return nil
}

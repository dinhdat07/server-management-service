package cdc

import (
	"context"
	"fmt"
)

type TopicHandler interface {
	Handle(ctx context.Context, value []byte) error
}

type RouterHandler struct {
	handlers map[string]TopicHandler
}

func NewRouterHandler(handlers map[string]TopicHandler) *RouterHandler {
	return &RouterHandler{
		handlers: handlers,
	}
}

func (h *RouterHandler) Handle(ctx context.Context, topic string, value []byte) error {
	handler, ok := h.handlers[topic]
	if !ok {
		return fmt.Errorf("unsupported topic: %s", topic)
	}

	return handler.Handle(ctx, value)
}

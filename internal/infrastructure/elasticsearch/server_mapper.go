package elasticsearch

import (
	"fmt"
	"time"
)

func ServerDocumentFromDebeziumAfter(after map[string]any) (*ServerDocument, error) {
	doc := &ServerDocument{}

	if id, ok := after["server_id"].(string); ok {
		doc.ID = id
	} else {
		return nil, fmt.Errorf("missing or invalid server_id")
	}

	if name, ok := after["server_name"].(string); ok {
		doc.ServerName = name
	}

	if ipv4, ok := after["ipv4"].(string); ok {
		doc.IPv4 = ipv4
	}

	if status, ok := after["current_status"].(string); ok {
		doc.CurrentStatus = status
	}

	if failuresFloat, ok := after["consecutive_failures"].(float64); ok {
		doc.ConsecutiveFailures = int(failuresFloat)
	}

	if createdAtMicro, ok := after["created_at"].(float64); ok {
		doc.CreatedAt = time.UnixMicro(int64(createdAtMicro)).UTC()
	}

	if updatedAtMicro, ok := after["updated_at"].(float64); ok {
		doc.UpdatedAt = time.UnixMicro(int64(updatedAtMicro)).UTC()
	}

	return doc, nil
}

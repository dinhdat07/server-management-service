package config

import "os"

type ElasticsearchConfig struct {
	URL        string
	UserIndex  string
	AuditIndex string
}

func LoadElasticsearchConfig() ElasticsearchConfig {
	url := os.Getenv("ELASTICSEARCH_URL")
	if url == "" {
		url = "http://elasticsearch:9200"
	}

	userIndex := os.Getenv("ELASTICSEARCH_USER_INDEX")
	if userIndex == "" {
		userIndex = "portal_users"
	}

	auditIndex := os.Getenv("ELASTICSEARCH_AUDIT_LOG_INDEX")
	if auditIndex == "" {
		auditIndex = "portal_users"
	}

	return ElasticsearchConfig{
		URL:        url,
		UserIndex:  userIndex,
		AuditIndex: auditIndex,
	}
}

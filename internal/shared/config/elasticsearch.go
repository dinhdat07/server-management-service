package config

import "os"

type ElasticsearchConfig struct {
	URL         string
	ServerIndex    string
	StatusLogIndex string
}

func LoadElasticsearchConfig() ElasticsearchConfig {
	url := os.Getenv("ELASTICSEARCH_URL")
	if url == "" {
		url = "http://elasticsearch:9200"
	}

	serverIndex := os.Getenv("ELASTICSEARCH_SERVER_INDEX")
	if serverIndex == "" {
		serverIndex = "sms_server_catalog"
	}

	statusLogIndex := os.Getenv("ELASTICSEARCH_STATUS_LOG_INDEX")
	if statusLogIndex == "" {
		statusLogIndex = "sms_status_logs"
	}

	return ElasticsearchConfig{
		URL:            url,
		ServerIndex:    serverIndex,
		StatusLogIndex: statusLogIndex,
	}
}

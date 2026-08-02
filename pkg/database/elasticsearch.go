package database

import (
	"fmt"
	"strings"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type ElasticsearchConfig struct {
	URLs     []string
	Username string
	Password string
}

func NewElasticsearch(cfg *viper.Viper, log *zap.Logger) (*elasticsearch.Client, error) {
	config := ElasticsearchConfig{
		URLs:     strings.Split(cfg.GetString("ELASTICSEARCH_URLS"), ","),
		Username: cfg.GetString("ELASTICSEARCH_USERNAME"),
		Password: cfg.GetString("ELASTICSEARCH_PASSWORD"),
	}

	// Configure Elasticsearch
	esCfg := elasticsearch.Config{
		Addresses: config.URLs,
		Username:  config.Username,
		Password:  config.Password,
	}

	// Create client
	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// Ping Elasticsearch
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to ping elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch returned error: %s", res.String())
	}

	log.Info("Elasticsearch connected successfully",
		zap.Strings("urls", config.URLs),
	)

	return client, nil
}

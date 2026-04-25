package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
)

type config struct {
	Addresses      []string
	Username       string
	Password       string
	IndexName      string
	CheckpointName string
	BatchSize      int
	PollInterval   time.Duration
}

type bulkResponse struct {
	Errors bool                     `json:"errors"`
	Items  []map[string]bulkItemRow `json:"items"`
}

type bulkItemRow struct {
	Status int             `json:"status"`
	Error  json.RawMessage `json:"error,omitempty"`
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		logger.Logf(logger.Error, "Config error: %v", err)
		os.Exit(1)
	}

	data.OpenDatabase()

	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	})
	if err != nil {
		logger.Logf(logger.Error, "Failed to create OpenSearch client: %v", err)
		os.Exit(1)
	}

	if err := ensureKeylogsIndex(context.Background(), client, cfg.IndexName); err != nil {
		logger.Logf(logger.Error, "Failed to ensure OpenSearch index: %v", err)
		os.Exit(1)
	}

	lastID, err := data.GetIndexerCheckpoint(cfg.CheckpointName)
	if err != nil {
		logger.Logf(logger.Error, "Failed to read checkpoint: %v", err)
		os.Exit(1)
	}
	logger.Logf(logger.Info, "Starting indexer at keylog_id=%d", lastID)

	for {
		docs, maxID, err := data.FetchKeylogDocumentsSince(lastID, cfg.BatchSize)
		if err != nil {
			logger.Logf(logger.Error, "Fetch error: %v", err)
			time.Sleep(cfg.PollInterval)
			continue
		}
		if len(docs) == 0 {
			time.Sleep(cfg.PollInterval)
			continue
		}

		if err := bulkIndex(context.Background(), client, cfg.IndexName, docs); err != nil {
			logger.Logf(logger.Error, "Bulk index failed: %v", err)
			time.Sleep(cfg.PollInterval)
			continue
		}

		if err := data.UpdateIndexerCheckpoint(cfg.CheckpointName, maxID); err != nil {
			logger.Logf(logger.Error, "Checkpoint update failed: %v", err)
			time.Sleep(cfg.PollInterval)
			continue
		}
		lastID = maxID
		logger.Logf(logger.Info, "Indexed %d docs (last_keylog_id=%d)", len(docs), lastID)

		if len(docs) < cfg.BatchSize {
			time.Sleep(cfg.PollInterval)
		}
	}
}

func bulkIndex(ctx context.Context, client *opensearch.Client, index string, docs []data.KeylogDocument) error {
	if len(docs) == 0 {
		return nil
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	for _, doc := range docs {
		action := map[string]map[string]string{
			"index": {
				"_index": index,
				"_id":    fmt.Sprintf("%d", doc.KeylogID),
			},
		}
		if err := enc.Encode(action); err != nil {
			return err
		}
		if err := enc.Encode(doc); err != nil {
			return err
		}
	}

	resp, err := client.Bulk(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bulk http %d: %s", resp.StatusCode, string(body))
	}

	var br bulkResponse
	if err := json.NewDecoder(resp.Body).Decode(&br); err != nil {
		return err
	}
	if !br.Errors {
		return nil
	}

	errCount := 0
	for _, item := range br.Items {
		for _, result := range item {
			if result.Status >= 300 {
				errCount++
			}
		}
	}
	if errCount > 0 {
		return fmt.Errorf("bulk indexing reported %d item errors", errCount)
	}
	return nil
}

func ensureKeylogsIndex(ctx context.Context, client *opensearch.Client, index string) error {
	existsResp, err := client.Indices.Exists([]string{index})
	if err != nil {
		return err
	}
	if existsResp != nil && existsResp.StatusCode == 200 {
		_ = existsResp.Body.Close()
		return nil
	}
	if existsResp != nil {
		_ = existsResp.Body.Close()
	}

	body := `{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0
  },
  "mappings": {
    "properties": {
      "keylog_id": { "type": "long" },
      "contents": { "type": "text" },
      "created_at": { "type": "date" },
      "uuid": { "type": "keyword" },
      "ip": { "type": "ip" },
      "tags": {
        "type": "nested",
        "properties": {
          "key": { "type": "keyword" },
          "value": { "type": "keyword" }
        }
      }
    }
  }
}`

	resp, err := client.Indices.Create(
		index,
		client.Indices.Create.WithContext(ctx),
		client.Indices.Create.WithBody(strings.NewReader(body)),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("index create http %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func loadConfig() (config, error) {
	addresses := splitCommaEnv("OPENSEARCH_HOSTS")
	if len(addresses) == 0 {
		address := strings.TrimSpace(os.Getenv("OPENSEARCH_URL"))
		if address == "" {
			address = "http://opensearch:9200"
		}
		addresses = []string{address}
	}

	batchSize, err := getenvInt("INDEXER_BATCH_SIZE", 500)
	if err != nil {
		return config{}, err
	}

	interval, err := getenvDuration("INDEXER_POLL_INTERVAL", 15*time.Second)
	if err != nil {
		return config{}, err
	}

	indexName := strings.TrimSpace(os.Getenv("OPENSEARCH_INDEX"))
	if indexName == "" {
		indexName = "keylogs"
	}

	checkpointName := strings.TrimSpace(os.Getenv("INDEXER_CHECKPOINT_NAME"))
	if checkpointName == "" {
		checkpointName = "keylogs"
	}

	return config{
		Addresses:      addresses,
		Username:       strings.TrimSpace(os.Getenv("OPENSEARCH_USERNAME")),
		Password:       os.Getenv("OPENSEARCH_PASSWORD"),
		IndexName:      indexName,
		CheckpointName: checkpointName,
		BatchSize:      batchSize,
		PollInterval:   interval,
	}, nil
}

func splitCommaEnv(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getenvInt(key string, def int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def, nil
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s=%q: %w", key, raw, err)
	}
	if val <= 0 {
		return 0, fmt.Errorf("%s must be > 0", key)
	}
	return val, nil
}

func getenvDuration(key string, def time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def, nil
	}
	val, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s=%q: %w", key, raw, err)
	}
	if val <= 0 {
		return 0, errors.New(key + " must be > 0")
	}
	return val, nil
}

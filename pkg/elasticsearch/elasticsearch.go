package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/putmapping"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

// Config contains Elasticsearch connection configuration
type Config struct {
	Host      string
	Username  string
	Password  string
	IndexName string
}

// Elastic wraps Elasticsearch TypedClient and provides generic operations
type Elastic struct {
	indexName string
	client    *elasticsearch.TypedClient
}

// SearchResult contains search results with metadata
type SearchResult struct {
	Total     int64             // Total number of hits
	MaxScore  float64           // Maximum relevance score
	Documents []json.RawMessage // Raw JSON documents
}

// NewElastic creates a new Elasticsearch client wrapper
func NewElastic(config Config) (*Elastic, error) {
	if config.IndexName == "" {
		return nil, fmt.Errorf("index name cannot be empty")
	}
	cfg := elasticsearch.Config{
		Addresses: []string{config.Host},
		Username:  config.Username,
		Password:  config.Password,
	}

	es, err := elasticsearch.NewTypedClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	return &Elastic{client: es, indexName: config.IndexName}, nil
}

// Prepare creates index with mapping or updates mapping if index exists
func (el *Elastic) Prepare(ctx context.Context, mappingByte []byte) error {
	var mapping types.TypeMapping

	if err := json.Unmarshal(mappingByte, &mapping); err != nil {
		return fmt.Errorf("failed to unmarshal mapping: %w", err)
	}

	_, err := el.client.Indices.Get(el.indexName).Do(ctx)
	if err != nil {
		// Index doesn't exist, create with mapping
		return el.createIndexWithMapping(ctx, &mapping)
	}

	// Index exists, update mapping
	return el.updateMapping(ctx, &mapping)
}

// Create indexes a new document with the given ID
func (el *Elastic) Create(ctx context.Context, id string, document interface{}) error {
	_, err := el.client.Index(el.indexName).
		Id(id).
		Request(document).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to create document with id %s: %w", id, err)
	}
	return nil
}

// Update updates an existing document
func (el *Elastic) Update(ctx context.Context, id string, document interface{}) error {
	_, err := el.client.Update(el.indexName, id).
		Doc(document).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to update document with id %s: %w", id, err)
	}
	return nil
}

// Delete removes a document by ID
func (el *Elastic) Delete(ctx context.Context, id string) error {
	_, err := el.client.Delete(el.indexName, id).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete document with id %s: %w", id, err)
	}
	return nil
}

// Get retrieves a document by ID and unmarshals it into the provided document
func (el *Elastic) Get(ctx context.Context, id string, document interface{}) error {
	res, err := el.client.Get(el.indexName, id).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to get document with id %s: %w", id, err)
	}

	if !res.Found {
		return fmt.Errorf("document with id %s not found", id)
	}

	if err := json.Unmarshal(res.Source_, document); err != nil {
		return fmt.Errorf("failed to unmarshal document: %w", err)
	}

	return nil
}

// Search performs a search query and returns results with metadata
func (el *Elastic) Search(ctx context.Context, req *search.Request) (*SearchResult, error) {
	if req.From != nil && *req.From < 0 {
		return nil, fmt.Errorf("offset cannot be negative")
	}
	if req.Size != nil && *req.Size <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}
	if req.Size != nil && *req.Size > 10000 {
		return nil, fmt.Errorf("limit cannot exceed 10000")
	}

	res, err := el.client.Search().
		Index(el.indexName).
		Request(req).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	result := &SearchResult{
		Total:     res.Hits.Total.Value,
		Documents: make([]json.RawMessage, 0, len(res.Hits.Hits)),
	}

	if res.Hits.MaxScore != nil {
		result.MaxScore = float64(*res.Hits.MaxScore)
	}

	for _, hit := range res.Hits.Hits {
		result.Documents = append(result.Documents, hit.Source_)
	}

	return result, nil
}

// SearchAndUnmarshal performs a search and unmarshals results directly into documents slice
// documents must be a pointer to a slice (e.g., *[]MyStruct)
func (el *Elastic) SearchAndUnmarshal(ctx context.Context, req *search.Request, documents interface{}) error {
	result, err := el.Search(ctx, req)
	if err != nil {
		return err
	}

	// Build JSON array manually to avoid double marshaling
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, source := range result.Documents {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.Write(source)
	}
	buf.WriteByte(']')

	if err := json.Unmarshal(buf.Bytes(), documents); err != nil {
		return fmt.Errorf("failed to unmarshal search results: %w", err)
	}

	return nil
}

// BulkIndex indexes multiple documents in a single request
// documents is a map of document ID to document
func (el *Elastic) BulkIndex(ctx context.Context, documents map[string]interface{}) error {
	if len(documents) == 0 {
		return nil
	}

	bulk := el.client.Bulk().Index(el.indexName)

	for id, doc := range documents {
		if err := bulk.CreateOp(types.CreateOperation{Id_: &id}, doc); err != nil {
			return fmt.Errorf("failed to add document %s to bulk: %w", id, err)
		}
	}

	res, err := bulk.Do(ctx)
	if err != nil {
		return fmt.Errorf("bulk index failed: %w", err)
	}

	if res.Errors {
		return fmt.Errorf("bulk index completed with errors")
	}

	return nil
}

// createIndexWithMapping creates a new index with the given mapping
func (el *Elastic) createIndexWithMapping(ctx context.Context, mp *types.TypeMapping) error {
	_, err := el.client.Indices.Create(el.indexName).
		Request(&create.Request{Mappings: mp}).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	return nil
}

// updateMapping updates the mapping of an existing index
func (el *Elastic) updateMapping(ctx context.Context, mp *types.TypeMapping) error {
	_, err := el.client.Indices.PutMapping(el.indexName).
		Request(&putmapping.Request{Properties: mp.Properties}).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to update mapping: %w", err)
	}
	return nil
}

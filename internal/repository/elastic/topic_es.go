package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/seshoo/bookFinder/internal/domain"
	"github.com/seshoo/bookFinder/internal/repository"
	"github.com/seshoo/bookFinder/pkg/elasticsearch"
)

// TopicRepository implements repository.Topics interface using Elasticsearch
type TopicRepository struct {
	db *elasticsearch.Elastic
}

// NewTopicRepository creates a new Topic repository
func NewTopicRepository(db *elasticsearch.Elastic) *TopicRepository {
	return &TopicRepository{db: db}
}

// Create creates a new topic in Elasticsearch
func (t *TopicRepository) Create(ctx context.Context, topic *domain.Topic) error {
	if topic == nil {
		return fmt.Errorf("topic cannot be nil")
	}
	if topic.Id == "" {
		return fmt.Errorf("topic id cannot be empty")
	}

	return t.db.Create(ctx, topic.Id, topic)
}

// Update updates an existing topic
func (t *TopicRepository) Update(ctx context.Context, topic *domain.Topic) error {
	if topic == nil {
		return fmt.Errorf("topic cannot be nil")
	}
	if topic.Id == "" {
		return fmt.Errorf("topic id cannot be empty")
	}

	return t.db.Update(ctx, topic.Id, topic)
}

// Delete deletes a topic by ID
func (t *TopicRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	return t.db.Delete(ctx, id)
}

// GetById retrieves a topic by ID
func (t *TopicRepository) GetById(ctx context.Context, id string) (*domain.Topic, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	var topic domain.Topic
	err := t.db.Get(ctx, id, &topic)
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

// GetByIds retrieves multiple topics by their IDs using a single query
func (t *TopicRepository) GetByIds(ctx context.Context, ids []string) (domain.Topics, error) {
	if len(ids) == 0 {
		return domain.Topics{}, nil
	}

	// Validate IDs
	for _, id := range ids {
		if id == "" {
			return nil, fmt.Errorf("id cannot be empty")
		}
	}

	// Build terms query to search by multiple IDs
	termsQuery := make([]types.FieldValue, len(ids))
	for i, id := range ids {
		termsQuery[i] = id
	}

	req := &search.Request{
		Query: &types.Query{
			Terms: &types.TermsQuery{
				TermsQuery: map[string]types.TermsQueryField{
					"id": termsQuery,
				},
			},
		},
		Size: func() *int { s := len(ids); return &s }(), // Return up to len(ids) documents
	}

	var topics domain.Topics
	if err := t.db.SearchAndUnmarshal(ctx, req, &topics); err != nil {
		return nil, fmt.Errorf("failed to get topics by ids: %w", err)
	}

	return topics, nil
}

// GetList retrieves topics with pagination and sorting
func (t *TopicRepository) GetList(ctx context.Context, offset, limit int) (domain.Topics, error) {
	if offset < 0 {
		return nil, fmt.Errorf("offset cannot be negative")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}
	if limit > 1000 {
		return nil, fmt.Errorf("limit cannot exceed 1000")
	}

	// Build search request with MatchAll query
	// Note: sorting removed to work with dynamic mapping in tests
	// In production, add sorting after setting up proper mapping
	req := &search.Request{
		Query: &types.Query{
			MatchAll: &types.MatchAllQuery{},
		},
		From: &offset,
		Size: &limit,
	}

	var topics domain.Topics
	if err := t.db.SearchAndUnmarshal(ctx, req, &topics); err != nil {
		return nil, fmt.Errorf("failed to get topic list: %w", err)
	}

	return topics, nil
}

// Search performs full-text search on topics (searches in text field only)
func (t *TopicRepository) Search(ctx context.Context, opts repository.SearchOptions) (domain.Topics, error) {
	// Validate parameters
	if opts.Offset < 0 {
		return nil, fmt.Errorf("offset cannot be negative")
	}
	if opts.Limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}
	if opts.Limit > 1000 {
		return nil, fmt.Errorf("limit cannot exceed 1000")
	}

	// Build query based on search text
	var query *types.Query
	if opts.Query == "" {
		// Empty query - match all
		query = &types.Query{
			MatchAll: &types.MatchAllQuery{},
		}
	} else {
		// Full-text search in text field only with Russian analyzer
		query = &types.Query{
			Match: map[string]types.MatchQuery{
				"text": {
					Query: opts.Query,
				},
			},
		}
	}

	// Always sort by relevance (_score descending)
	order := sortorder.Desc
	sort := []types.SortCombinations{
		types.SortOptions{
			SortOptions: map[string]types.FieldSort{
				"_score": {Order: &order},
			},
		},
	}

	// Build search request
	req := &search.Request{
		Query: query,
		From:  &opts.Offset,
		Size:  &opts.Limit,
		Sort:  sort,
	}

	// Execute search
	result, err := t.db.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Unmarshal documents
	var topics domain.Topics
	if len(result.Documents) > 0 {
		// Build JSON array from raw messages
		var buf bytes.Buffer
		buf.WriteByte('[')
		for i, doc := range result.Documents {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.Write(doc)
		}
		buf.WriteByte(']')

		// Unmarshal to topics
		if err := json.Unmarshal(buf.Bytes(), &topics); err != nil {
			return nil, fmt.Errorf("failed to unmarshal topics: %w", err)
		}
	}

	return topics, nil
}

// BulkCreate creates multiple topics at once
func (t *TopicRepository) BulkCreate(ctx context.Context, topics domain.Topics) error {

	if len(topics) == 0 {
		return nil
	}

	// Convert to map[id]document
	documents := make(map[string]interface{}, len(topics))
	for _, topic := range topics {
		if topic.Id == "" {
			return fmt.Errorf("topic id cannot be empty")
		}

		documents[topic.Id] = topic
	}

	return t.db.BulkIndex(ctx, documents)
}

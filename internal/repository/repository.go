package repository

import (
	"context"

	"github.com/seshoo/bookFinder/internal/domain"
)

// Topics defines repository interface for Topic operations
type Topics interface {
	// GetById retrieves a topic by ID
	GetById(ctx context.Context, id string) (*domain.Topic, error)

	// GetByIds retrieves multiple topics by their IDs
	GetByIds(ctx context.Context, ids []string) (domain.Topics, error)

	// GetList retrieves topics with pagination
	GetList(ctx context.Context, offset, limit int) (domain.Topics, error)

	// Create creates a new topic
	Create(ctx context.Context, topic *domain.Topic) error

	// Update updates an existing topic
	Update(ctx context.Context, topic *domain.Topic) error

	// Delete deletes a topic by ID
	Delete(ctx context.Context, id string) error

	// Search performs full-text search with options
	Search(ctx context.Context, opts SearchOptions) (domain.Topics, error)

	// BulkCreate creates multiple topics at once
	BulkCreate(ctx context.Context, topics domain.Topics) error
}

// SearchOptions contains search parameters
type SearchOptions struct {
	Query  string // Search query (searches in text field only)
	Offset int    // Pagination offset
	Limit  int    // Pagination limit
}

// Repositories aggregates all repository interfaces
type Repositories interface {
	Topic() Topics
}

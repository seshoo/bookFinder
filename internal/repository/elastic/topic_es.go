package elastic

import (
	"context"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/seshoo/bookFinder/internal/domain"
	"github.com/seshoo/bookFinder/pkg/elasticsearch"
)

type TopicRepository struct {
	db *elasticsearch.Elastic
}

func NewTopicRepository(db *elasticsearch.Elastic) *TopicRepository {
	return &TopicRepository{db: db}
}

func (t TopicRepository) GetList(ctx context.Context, offset int, limit int) ([]domain.Topic, error) {
	query := &types.Query{
		MatchAll: &types.MatchAllQuery{},
	}
	var topics []domain.Topic
	err := t.db.Search(ctx, query, offset, limit, &topics)
	if err != nil {
		return nil, err
	}

	return topics, nil
}

func (t TopicRepository) GetById(ctx context.Context, id string) (*domain.Topic, error) {
	var topic domain.Topic
	err := t.db.Get(ctx, id, &topic)
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

package elastic

import (
	"github.com/seshoo/bookFinder/internal/repository"
	"github.com/seshoo/bookFinder/pkg/elasticsearch"
)

type ElasticRepository struct {
	topic repository.Topics
}

func NewElasticRepository(db *elasticsearch.Elastic) *ElasticRepository {
	return &ElasticRepository{
		topic: NewTopicRepository(db),
	}
}

func (e *ElasticRepository) Topic() repository.Topics {
	return e.topic
}

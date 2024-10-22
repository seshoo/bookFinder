package elastic_test

import (
	"context"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/seshoo/bookFinder/internal/domain"
	"github.com/seshoo/bookFinder/internal/repository/elastic"
	"github.com/seshoo/bookFinder/pkg/elasticsearch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var mappingString = `{
  "mappings": {
    "properties": {
      "id": {
        "type": "keyword"
      },
      "title": {
        "type": "text",
        "analyzer": "russian"
      },
      "link": {
        "type": "keyword"
      },
      "text": {
        "type": "text",
        "analyzer": "russian"
      }
    }
  }
}`

func TestTopicRepository(t *testing.T) {
	var (
		err        error
		esUsername = "test"
		esPassword = "test"
	)

	esHost, teardown := startTestContainer(t, esUsername, esPassword)
	defer teardown()

	esClient := getClient(t, esHost, esUsername, esPassword)

	baseIndexName := "test_topic_index"

	t.Run("GetById return topic if it exists", func(t *testing.T) {
		testIndexName := baseIndexName + "_get_by_id_exists"
		createIndex(t, esClient, testIndexName)
		defer deleteIndex(t, esClient, testIndexName)
		//setMapping(t, esClient, testIndexName, mappingString)

		expectedTopic := &domain.Topic{
			Id:    "1",
			Title: "test title",
			Link:  "test link",
			Text:  "test text",
		}

		_, err := esClient.Index(testIndexName).
			Id(expectedTopic.Id).
			Request(expectedTopic).
			Do(context.TODO())
		if err != nil {
			t.Fatalf("failed to index document: %s", err)
		}

		// Обновляем имя индекса в db для этого теста
		db, _ := elasticsearch.NewElastic(elasticsearch.Config{
			Host:      esHost,
			Username:  esUsername,
			Password:  esPassword,
			IndexName: testIndexName,
		})
		topicRepository := elastic.NewTopicRepository(db)

		actualTopic, err := topicRepository.GetById(context.Background(), expectedTopic.Id)

		assert.Equal(t, expectedTopic, actualTopic)
		require.NoError(t, err)
	})

	t.Run("GetById return topic if it not exists", func(t *testing.T) {
		testIndexName := baseIndexName + "_get_by_id_not_exists"
		createIndex(t, esClient, testIndexName)
		defer deleteIndex(t, esClient, testIndexName)
		//setMapping(t, esClient, testIndexName, mappingString)

		db, _ := elasticsearch.NewElastic(elasticsearch.Config{
			Host:      esHost,
			Username:  esUsername,
			Password:  esPassword,
			IndexName: testIndexName,
		})
		topicRepository := elastic.NewTopicRepository(db)

		actualTopic, err := topicRepository.GetById(context.Background(), "1")

		assert.Nil(t, actualTopic)
		require.Error(t, err)
	})

	t.Run("GetList return topics", func(t *testing.T) {
		testIndexName := baseIndexName + "_get_list"
		createIndex(t, esClient, testIndexName)
		defer deleteIndex(t, esClient, testIndexName)
		//setMapping(t, esClient, testIndexName, mappingString)

		db, _ := elasticsearch.NewElastic(elasticsearch.Config{
			Host:      esHost,
			Username:  esUsername,
			Password:  esPassword,
			IndexName: testIndexName,
		})
		topicRepository := elastic.NewTopicRepository(db)

		bulk := esClient.Bulk().Index(testIndexName).Refresh(refresh.True)

		docs := []*domain.Topic{
			{
				Id:    "1",
				Title: "test title 1",
				Link:  "test link 1",
				Text:  "test text 1",
			},
			{
				Id:    "2",
				Title: "test title 2",
				Link:  "test link 2",
				Text:  "test text 2",
			},
			{
				Id:    "3",
				Title: "test title 3",
				Link:  "test link 3",
				Text:  "test text 3",
			},
		}

		for _, doc := range docs {
			err = bulk.CreateOp(types.CreateOperation{Id_: &doc.Id}, doc)
			if err != nil {
				t.Fatalf("failed to create bulk operation: %s", err)
			}
		}

		_, err = bulk.Do(context.Background())
		if err != nil {
			t.Fatalf("failed to bulk index documents: %s", err)
		}

		topics, err := topicRepository.GetList(context.Background(), 0, 2)

		require.NoError(t, err)
		assert.Len(t, topics, 2)
		assert.Contains(t, topics, *docs[0])
		assert.Contains(t, topics, *docs[1])

		topics, err = topicRepository.GetList(context.Background(), 1, 2)

		require.NoError(t, err)
		assert.Len(t, topics, 2)
		assert.Contains(t, topics, *docs[1])
		assert.Contains(t, topics, *docs[2])

		topics, err = topicRepository.GetList(context.Background(), 2, 2)

		require.NoError(t, err)
		assert.Len(t, topics, 1)
		assert.Contains(t, topics, *docs[2])
	})
}

package elastic_test

import (
	"context"
	"testing"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/seshoo/bookFinder/internal/domain"
	"github.com/seshoo/bookFinder/internal/repository"
	"github.com/seshoo/bookFinder/internal/repository/elastic"
	"github.com/seshoo/bookFinder/pkg/elasticsearch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	t.Run("GetByIds returns topics by list of IDs", func(t *testing.T) {
		testIndexName := baseIndexName + "_get_by_ids"
		createIndex(t, esClient, testIndexName)
		defer deleteIndex(t, esClient, testIndexName)

		db, err := elasticsearch.NewElastic(elasticsearch.Config{
			Host:      esHost,
			Username:  esUsername,
			Password:  esPassword,
			IndexName: testIndexName,
		})
		require.NoError(t, err)

		topicRepository := elastic.NewTopicRepository(db)

		// Create test topics
		testTopics := domain.Topics{
			{Id: "1", Title: "Topic 1", Link: "http://example.com/1", Text: "Text 1"},
			{Id: "2", Title: "Topic 2", Link: "http://example.com/2", Text: "Text 2"},
			{Id: "3", Title: "Topic 3", Link: "http://example.com/3", Text: "Text 3"},
			{Id: "4", Title: "Topic 4", Link: "http://example.com/4", Text: "Text 4"},
			{Id: "5", Title: "Topic 5", Link: "http://example.com/5", Text: "Text 5"},
		}

		err = topicRepository.BulkCreate(context.Background(), testTopics)
		require.NoError(t, err)

		// Wait for indexing
		_, err = esClient.Indices.Refresh().Index(testIndexName).Do(context.Background())
		require.NoError(t, err)

		// Test 1: Get existing topics by IDs
		topics, err := topicRepository.GetByIds(context.Background(), []string{"1", "3", "5"})
		require.NoError(t, err)
		assert.Len(t, topics, 3)

		// Verify correct topics returned
		foundIds := make(map[string]bool)
		for _, topic := range topics {
			foundIds[topic.Id] = true
		}
		assert.True(t, foundIds["1"])
		assert.True(t, foundIds["3"])
		assert.True(t, foundIds["5"])

		// Test 2: Get mix of existing and non-existing IDs
		topics, err = topicRepository.GetByIds(context.Background(), []string{"1", "99", "3", "100"})
		require.NoError(t, err)
		assert.Len(t, topics, 2, "should only return existing topics")

		foundIds = make(map[string]bool)
		for _, topic := range topics {
			foundIds[topic.Id] = true
		}
		assert.True(t, foundIds["1"])
		assert.True(t, foundIds["3"])
		assert.False(t, foundIds["99"])
		assert.False(t, foundIds["100"])

		// Test 3: Empty ID list
		topics, err = topicRepository.GetByIds(context.Background(), []string{})
		require.NoError(t, err)
		assert.Len(t, topics, 0)

		// Test 4: All non-existing IDs
		topics, err = topicRepository.GetByIds(context.Background(), []string{"99", "100", "101"})
		require.NoError(t, err)
		assert.Len(t, topics, 0)
	})

	t.Run("Search with Russian morphology analyzer", func(t *testing.T) {
		testIndexName := baseIndexName + "_search_russian"
		createIndex(t, esClient, testIndexName)
		defer deleteIndex(t, esClient, testIndexName)

		setMapping(t, esClient, testIndexName, mappingString)

		db, err := elasticsearch.NewElastic(elasticsearch.Config{
			Host:      esHost,
			Username:  esUsername,
			Password:  esPassword,
			IndexName: testIndexName,
		})
		require.NoError(t, err)

		topicRepository := elastic.NewTopicRepository(db)

		topics := domain.Topics{
			{
				Id:    "1",
				Title: "Книга о программировании",
				Link:  "http://example.com/1",
				Text:  "Эта книга рассказывает о программировании на Go",
			},
			{
				Id:    "2",
				Title: "Статья о котах",
				Link:  "http://example.com/2",
				Text:  "В этой статье мы говорим о котах и кошках",
			},
			{
				Id:    "3",
				Title: "Учебник по базам данных",
				Link:  "http://example.com/3",
				Text:  "Книги по базам данных помогают понять SQL",
			},
		}

		// Индексируем все топики
		err = topicRepository.BulkCreate(context.Background(), topics)
		require.NoError(t, err)

		_, err = esClient.Indices.Refresh().Index(testIndexName).Do(context.Background())
		require.NoError(t, err)

		// Табличные тесты для проверки русской морфологии
		testCases := []struct {
			name          string
			query         string
			expectedTotal int64
			expectedIds   []string // ID топиков, которые должны быть найдены
			notFoundIds   []string // ID топиков, которые НЕ должны быть найдены
		}{
			{
				name:          "поиск 'книга' находит 'книга' и 'книги'",
				query:         "книга",
				expectedTotal: 2,
				expectedIds:   []string{"1", "3"},
				notFoundIds:   []string{"2"},
			},
			{
				name:          "поиск 'кот' находит 'котах' и 'кошках'",
				query:         "кот",
				expectedTotal: 1,
				expectedIds:   []string{"2"},
				notFoundIds:   []string{"1", "3"},
			},
			{
				name:          "поиск 'программирование' находит 'программировании'",
				query:         "программирование",
				expectedTotal: 1,
				expectedIds:   []string{"1"},
				notFoundIds:   []string{"2", "3"},
			},
			{
				name:          "поиск 'база' находит 'базам'",
				query:         "база",
				expectedTotal: 1,
				expectedIds:   []string{"3"},
				notFoundIds:   []string{"1", "2"},
			},
			{
				name:          "поиск несуществующего слова",
				query:         "несуществующееслово",
				expectedTotal: 0,
				expectedIds:   []string{},
				notFoundIds:   []string{"1", "2", "3"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				topics, err := topicRepository.Search(context.Background(), repository.SearchOptions{
					Query:  tc.query,
					Offset: 0,
					Limit:  10,
				})

				require.NoError(t, err)
				assert.Equal(t, int(tc.expectedTotal), len(topics), "количество найденных топиков не совпадает")

				// Собираем ID найденных топиков
				foundIds := make(map[string]bool)
				for _, topic := range topics {
					foundIds[topic.Id] = true
				}

				// Проверяем, что все ожидаемые топики найдены
				for _, expectedId := range tc.expectedIds {
					assert.True(t, foundIds[expectedId], "топик %s должен быть найден", expectedId)
				}

				// Проверяем, что незапланированные топики не найдены
				for _, notFoundId := range tc.notFoundIds {
					assert.False(t, foundIds[notFoundId], "топик %s не должен быть найден", notFoundId)
				}
			})
		}
	})
}

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/seshoo/bookFinder/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopicSerialization(t *testing.T) {
	topic := domain.Topic{
		Id:    "123",
		Title: "Тестовая тема",
		Link:  "http://example.com",
		Text:  "Текст темы",
	}

	// Сериализация в JSON
	data, err := json.Marshal(topic)
	require.NoError(t, err, "Ошибка сериализации")

	// Десериализация из JSON
	var newTopic domain.Topic
	err = json.Unmarshal(data, &newTopic)
	require.NoError(t, err, "Ошибка десериализации")

	// Проверка полей
	assert.Equal(t, topic.Id, newTopic.Id, "Id должны совпадать")
	assert.Equal(t, topic.Title, newTopic.Title, "Title должны совпадать")
	assert.Equal(t, topic.Link, newTopic.Link, "Link должны совпадать")
	assert.Equal(t, topic.Text, newTopic.Text, "Text должны совпадать")
}

func TestTopicJSONTags(t *testing.T) {
	// Проверка соответствия JSON тегов
	topic := domain.Topic{
		Id:    "123",
		Title: "Заголовок",
		Link:  "http://test.com",
		Text:  "Содержание",
	}

	data, err := json.Marshal(topic)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"id":"123"`)
	assert.Contains(t, jsonStr, `"title":"Заголовок"`)
	assert.Contains(t, jsonStr, `"link":"http://test.com"`)
	assert.Contains(t, jsonStr, `"text":"Содержание"`)
}

func TestTopicFields(t *testing.T) {
	// Создаем тестовую структуру
	topic := domain.Topic{
		Id:    "test-id",
		Title: "Тестовый заголовок",
		Link:  "http://example.org",
		Text:  "Тестовый текст",
	}

	// Проверяем каждое поле
	assert.NotEmpty(t, topic.Id)
	assert.NotEmpty(t, topic.Title)
	assert.NotEmpty(t, topic.Link)
	assert.NotEmpty(t, topic.Text)

	assert.Equal(t, "test-id", topic.Id)
	assert.Equal(t, "Тестовый заголовок", topic.Title)
	assert.Equal(t, "http://example.org", topic.Link)
	assert.Equal(t, "Тестовый текст", topic.Text)
}

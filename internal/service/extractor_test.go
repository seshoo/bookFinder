package service_test

import (
	"github.com/seshoo/bookFinder/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractorService_Get_ValidContent(t *testing.T) {
	extractor := service.NewExtractorService()

	content := `<html><head><title>Test Title</title></head><body><div class="sp-body">Test Text</div></body></html>`
	expectedTitle := "Test Title"
	expectedText := "Test Text"

	title, text, err := extractor.Get(content)

	assert.NoError(t, err)
	assert.Equal(t, expectedTitle, title)
	assert.Equal(t, expectedText, text)
}

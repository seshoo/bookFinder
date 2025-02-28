package service_test

import (
	"fmt"
	"github.com/seshoo/bookFinder/internal/service"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientService_Do_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, `<html><body>Test Content</body></html>`)
	}))
	defer mockServer.Close()

	clientService := service.NewClientService(http.DefaultClient)
	content, err := clientService.Do(mockServer.URL)

	assert.NoError(t, err)
	assert.Contains(t, content, "Test Content")
}

func TestClientService_Do_Error(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	clientService := service.NewClientService(http.DefaultClient)
	content, err := clientService.Do(mockServer.URL)

	assert.Error(t, err)
	assert.Empty(t, content)
}

func TestClientService_Do_NonUTF8Charset(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=iso-8859-1")
		fmt.Fprintln(w, `<html><body>Test Content</body></html>`)
	}))
	defer mockServer.Close()

	clientService := service.NewClientService(http.DefaultClient)
	content, err := clientService.Do(mockServer.URL)

	assert.NoError(t, err)
	assert.Contains(t, content, "Test Content")
}

func TestClientService_Do_NonCharsetHeader(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, `<html><body>Test Content</body></html>`)
	}))
	defer mockServer.Close()

	clientService := service.NewClientService(http.DefaultClient)
	content, err := clientService.Do(mockServer.URL)

	assert.NoError(t, err)
	assert.Contains(t, content, "Test Content")
}

func TestClientService_Do_InvalidCharset(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=abc")
		fmt.Fprintln(w, `<html><body>Test Content</body></html>`)
	}))
	defer mockServer.Close()

	clientService := service.NewClientService(http.DefaultClient)
	content, err := clientService.Do(mockServer.URL)

	assert.Error(t, err)
	assert.Empty(t, content)
}

func TestClientService_Do_HttpGetError(t *testing.T) {
	// Create a mock server that simulates an error response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	clientService := service.NewClientService(http.DefaultClient)
	content, err := clientService.Do(mockServer.URL)

	assert.Error(t, err)
	assert.Empty(t, content)
}

type MockHttpClient struct {
	mock.Mock
}

func (m *MockHttpClient) Get(url string) (*http.Response, error) {
	args := m.Called(url)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*http.Response), args.Error(1)
}

func TestClientService_Do_HttpClientReturnError(t *testing.T) {
	mockHttpClient := new(MockHttpClient)
	clientService := service.NewClientService(mockHttpClient)

	mockHttpClient.On("Get", "http://example.com").Return(
		nil,
		fmt.Errorf("http get error"),
	)

	content, err := clientService.Do("http://example.com")

	assert.Error(t, err)
	assert.Empty(t, content)
}

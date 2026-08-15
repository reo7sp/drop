package main

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisIntegration(t *testing.T) {
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		t.Skip("TEST_REDIS_URL is not set")
	}

	ctx := context.Background()
	client, err := initRedis(ctx, redisURL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.FlushDB(ctx).Err())
	t.Cleanup(func() { _ = client.FlushDB(ctx).Err() })

	tmpl := template.Must(template.ParseGlob("templates/*"))
	handler := newHandler(client, tmpl)
	form := url.Values{"text": {"integration drop"}}
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusSeeOther, recorder.Code)
	assert.Equal(t, "/", recorder.Header().Get("Location"))

	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "integration drop")
}

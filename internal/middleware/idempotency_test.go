package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rizky/smart-grant/pkg/idempotency"
	"github.com/stretchr/testify/assert"
)

func TestIdempotency_SkipsNonMutating(t *testing.T) {
	store := idempotency.NewStore(nil)
	handler := Idempotency(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("Idempotency-Key", "key-123")
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIdempotency_NoKeyPassesThrough(t *testing.T) {
	store := idempotency.NewStore(nil)
	called := false
	handler := Idempotency(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"success":true}`))
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"name":"test"}`)))
	handler.ServeHTTP(w, r)

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestIdempotency_WithKeyCachesResponse(t *testing.T) {
	store := idempotency.NewStore(nil)
	called := 0
	handler := Idempotency(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"success":true}`))
	}))

	body := bytes.NewReader([]byte(`{"name":"test"}`))
	r := httptest.NewRequest("POST", "/test", body)
	r.Header.Set("Idempotency-Key", "unique-key-1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	body2 := bytes.NewReader([]byte(`{"name":"test"}`))
	r2 := httptest.NewRequest("POST", "/test", body2)
	r2.Header.Set("Idempotency-Key", "unique-key-1")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, r2)

	assert.Equal(t, 2, called, "nil store = no caching, key acts as passthrough")
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, http.StatusCreated, w2.Code)
}

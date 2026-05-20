package middleware

import (
	"bytes"
	"net/http"

	"github.com/rizky/smart-grant/pkg/idempotency"
)

func Idempotency(store *idempotency.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			if store.IsProcessed(r.Context(), key) {
				cached, err := store.Get(r.Context(), key)
				if err == nil && cached != nil {
					for k, v := range cached.Headers {
						w.Header().Set(k, v)
					}
					w.WriteHeader(cached.StatusCode)
					w.Write(cached.Body)
					return
				}
			}

			rec := &recorder{ResponseWriter: w, body: &bytes.Buffer{}}
			next.ServeHTTP(rec, r)

			if rec.statusCode < 500 {
				store.Set(r.Context(), key, &idempotency.CachedResponse{
					StatusCode: rec.statusCode,
					Body:       rec.body.Bytes(),
					Headers: map[string]string{
						"Content-Type": w.Header().Get("Content-Type"),
					},
				})
			}
		})
	}
}

type recorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (r *recorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *recorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

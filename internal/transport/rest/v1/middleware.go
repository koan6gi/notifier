package v1

import (
	"context"
	"net/http"

	"github.com/koan6gi/notifier/pkg/logger"
)

type Updater interface {
	Update(ctx context.Context)
}

func LoggingMiddleware(next http.Handler, lg *logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := r.WithContext(logger.WithContext(r.Context(), lg))

		next.ServeHTTP(w, req)
	})
}

func UpdateMiddleware(next http.Handler, u Updater) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u.Update(r.Context())

		next.ServeHTTP(w, r)
	})
}

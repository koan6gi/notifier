package v1

import (
	"net/http"

	"github.com/koan6gi/notifier/pkg/logger"
)

func LoggingMiddleware(next http.Handler, lg *logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := r.WithContext(logger.WithContext(r.Context(), lg))

		next.ServeHTTP(w, req)
	})
}

package v1

import (
	"encoding/json"
	"net/http"

	"github.com/koan6gi/notifier/internal/domain"
	"github.com/koan6gi/notifier/pkg/logger"

	"go.uber.org/zap"
)

type Service interface {
	ListResponses() []domain.Response
	LastResponse() *domain.Response
}

type Handlers struct {
	svc Service
}

func NewHandlers(svc Service) *Handlers {
	return &Handlers{
		svc: svc,
	}
}

func (h *Handlers) Last(w http.ResponseWriter, r *http.Request) {
	resp := h.svc.LastResponse()

	lg, _ := logger.FromContext(r.Context())

	enc := json.NewEncoder(w)

	enc.SetIndent("", "  ")

	if err := enc.Encode(resp); err != nil {
		lg.Error("last: failed write json", zap.Error(err))
		return
	}
}

func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	resp := h.svc.ListResponses()

	lg, _ := logger.FromContext(r.Context())

	enc := json.NewEncoder(w)

	enc.SetIndent("", "  ")

	if err := enc.Encode(resp); err != nil {
		lg.Error("list: failed write json", zap.Error(err))
		return
	}
}

package repository

import (
	"sync"
	"time"

	"github.com/koan6gi/notifier/internal/domain"
)

const (
	maxItemCount = 20
)

type Repository struct {
	st []domain.Response
	mu sync.RWMutex
}

func New() *Repository {
	return &Repository{
		mu: sync.RWMutex{},
	}
}

func (r *Repository) refreshRepo() {
	if len(r.st) <= maxItemCount {
		return
	}

	newStorage := make([]domain.Response, maxItemCount)

	copy(newStorage, r.st[len(r.st)-maxItemCount:])

	r.st = newStorage
}

func (r *Repository) AddResponse(items []domain.Item) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refreshRepo()

	r.st = append(r.st, domain.Response{
		Items:  items,
		Moment: time.Now(),
	})
}

func (r *Repository) ListResponses() []domain.Response {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rr := make([]domain.Response, len(r.st))

	copy(rr, r.st)

	return rr
}

func (r *Repository) LastResponse() *domain.Response {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.st) == 0 {
		return nil
	}

	resp := &domain.Response{
		Moment: r.st[len(r.st)-1].Moment,
		Items:  make([]domain.Item, 0),
	}

	for _, item := range r.st[len(r.st)-1].Items {
		c, ok := item.Count.(float64)
		if !ok || c > 0 {
			resp.Items = append(resp.Items, item)
		}
	}

	return resp
}

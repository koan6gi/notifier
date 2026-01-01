package service

import "github.com/koan6gi/notifier/internal/domain"

type Repository interface {
	ListResponses() []domain.Response
	LastResponse() *domain.Response
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) ListResponses() []domain.Response {
	return s.repo.ListResponses()
}

func (s *Service) LastResponse() *domain.Response {
	return s.repo.LastResponse()
}

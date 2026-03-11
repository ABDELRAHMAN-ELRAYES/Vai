package ai

import "context"

type Service struct {
	repo   *Repository
	client *Client
}

func NewService(repo *Repository, client *Client) *Service {
	return &Service{
		repo:   repo,
		client: client,
	}
}

func (s *Service) Generate(ctx context.Context, prompt string) (<-chan string, <-chan error) {
	return s.client.GenerateStream(ctx, prompt)
}

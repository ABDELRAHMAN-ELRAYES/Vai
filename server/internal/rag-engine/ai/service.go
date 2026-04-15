package ai

import (
	"context"
	"strings"
)


type IService interface {
	Generate(ctx context.Context, prompt string) (<-chan string, <-chan error)
	CollectTokens(tokenChan <-chan string, errChan <-chan error) (<-chan string, <-chan string, <-chan error)
	EmbedBatch(ctx context.Context, input []string) ([][]float32, error)
}

type Service struct {
	client *Client
}

func NewService(client *Client) *Service {
	return &Service{
		client: client,
	}
}

func (service *Service) Generate(ctx context.Context, prompt string) (<-chan string, <-chan error) {
	return service.client.GenerateStream(ctx, prompt)
}

// Collect the response tokens stream and return a full reply string through a channel
func (service *Service) CollectTokens(tokenChan <-chan string, errChan <-chan error) (<-chan string, <-chan string, <-chan error) {

	tokenStream := make(chan string)
	errStream := make(chan error, 1)
	replyChan := make(chan string, 1)

	go func() {
		defer close(tokenStream)
		defer close(errStream)

		var fullReply strings.Builder

		for {
			select {
			case token, ok := <-tokenChan:
				if !ok {
					replyChan <- strings.TrimSpace(fullReply.String())
					close(replyChan)
					return
				}
				fullReply.WriteString(token)
				tokenStream <- token
			case err, ok := <-errChan:
				if !ok {
					continue
				}
				if err != nil {
					replyChan <- strings.TrimSpace(fullReply.String())
					close(replyChan)
					errStream <- err
					return
				}
			}
		}
	}()

	return replyChan, tokenStream, errStream
}

func (service *Service) EmbedBatch(ctx context.Context, input []string) ([][]float32, error) {
	return service.client.EmbedBatch(ctx, input)
}

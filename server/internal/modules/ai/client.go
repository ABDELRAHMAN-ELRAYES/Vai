package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/app"
)

type Client struct {
	baseURL string
	http    *http.Client // responsible for sending requests and recieving responses
	app     *app.Application
}

func NewClient(app *app.Application, baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{},
		app:     app,
	}
}

func (c *Client) GenerateStream(ctx context.Context, prompt string) (<-chan string, <-chan error) {
	// Stream Tokens view channel
	tokenStream := make(chan string)
	errStream := make(chan error, 1)

	go func() {
		defer close(tokenStream)
		defer close(errStream)

		reqBody := GenerateRequest{
			Model:  c.app.Config.AI.Name,
			Prompt: prompt,
			Stream: true,
		}

		data, _ := json.Marshal(reqBody)

		// Make a request to the AI Model
		req, err := http.NewRequestWithContext(
			ctx,
			"POST",
			c.baseURL+"/api/generate",
			bytes.NewBuffer(data),
		)

		if err != nil {
			errStream <- err
			return
		}

		req.Header.Set("Content-Type", "application/json")

		// 1. opens a TCP connection
		// 2. Sends the request to the ai model server
		// 3. waits for the server to respond
		// 4. return the response metadata
		// 5. provide a Body stream to the response to read the response as tokens not waiting for the full response to finish
		resp, err := c.http.Do(req)
		if err != nil {
			errStream <- err
			return
		}

		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)

		for {

			var chunk OllamaChunk
			// Decode the token returned from AI Model and send it throw the token channel
			if err := decoder.Decode(&chunk); err != nil {
				errStream <- err
				return
			}

			if chunk.Response != "" || chunk.Thinking != "" {
				tokenStream <- chunk.Response + chunk.Thinking
			}

			if chunk.Done {
				return
			}
		}
	}()

	return tokenStream, errStream
}

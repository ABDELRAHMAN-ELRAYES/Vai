package ai_test

import (
	"testing"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/rag-engine/ai"
	"github.com/stretchr/testify/assert"
)

func TestCollectTokens(t *testing.T) {
	service := ai.NewService(nil)

	t.Run("collect tokens successfully", func(t *testing.T) {
		tokenChan := make(chan string, 3)
		errChan := make(chan error, 1)

		tokenChan <- "Hello"
		tokenChan <- " "
		tokenChan <- "World"
		close(tokenChan)
		close(errChan)

		replyChan, tokenStream, errStream := service.CollectTokens(tokenChan, errChan)

		// Check token stream
		tokens := []string{}
		for token := range tokenStream {
			tokens = append(tokens, token)
		}
		assert.Equal(t, []string{"Hello", " ", "World"}, tokens)

		// Check full reply
		reply := <-replyChan
		assert.Equal(t, "Hello World", reply)

		// Check error stream
		err, ok := <-errStream
		assert.False(t, ok)
		assert.NoError(t, err)
	})

	t.Run("collect tokens with error", func(t *testing.T) {
		tokenChan := make(chan string, 2)
		errChan := make(chan error, 1)

		tokenChan <- "Hello"
		errChan <- assert.AnError
		close(tokenChan)
		close(errChan)

		replyChan, _, errStream := service.CollectTokens(tokenChan, errChan)

		err := <-errStream
		assert.Error(t, err)

		reply := <-replyChan
		assert.Equal(t, "Hello", reply)
	})
}

func TestRenderPrompt(t *testing.T) {
	err := ai.LoadPrompts()
	assert.NoError(t, err)

	t.Run("render query prompt", func(t *testing.T) {
		data := struct{ Text string }{Text: "test query"}
		prompt, err := ai.RenderPrompt(ai.EmbedQueryPrompt, data)
		assert.NoError(t, err)
		assert.Contains(t, prompt, "test query")
	})
}

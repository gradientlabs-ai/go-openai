package openai_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gradientlabs-ai/go-openai"
	"github.com/gradientlabs-ai/go-openai/internal/test/checks"
)

func TestCompletionRequestOmitEmpty(t *testing.T) {
	data, err := json.Marshal(openai.CompletionRequest{Model: "davinci-002"})
	checks.NoError(t, err)

	const expected = `{"model":"davinci-002"}`
	if string(data) != expected {
		t.Errorf("expected JSON with all empty fields to be %v but was %v", expected, string(data))
	}
}

func TestCompletionRequestPreservesZeroTemperature(t *testing.T) {
	var temperature float32
	data, err := json.Marshal(openai.CompletionRequest{Model: "davinci-002", Temperature: &temperature})
	checks.NoError(t, err)

	// An explicit zero temperature must be sent to override the API default.
	const expected = `{"model":"davinci-002","temperature":0}`
	if string(data) != expected {
		t.Errorf("expected JSON with zero temperature to be %v but was %v", expected, string(data))
	}
}

func TestCompletionsWrongModel(t *testing.T) {
	config := openai.DefaultConfig("whatever")
	config.BaseURL = "http://localhost/v1"
	client := openai.NewClientWithConfig(config)

	_, err := client.CreateCompletion(
		context.Background(),
		openai.CompletionRequest{
			MaxTokens: 5,
			Model:     openai.GPT3Dot5Turbo,
		},
	)
	if !errors.Is(err, openai.ErrCompletionUnsupportedModel) {
		t.Fatalf("CreateCompletion should return ErrCompletionUnsupportedModel, but returned: %v", err)
	}
}

func TestCompletionWithStream(t *testing.T) {
	config := openai.DefaultConfig("whatever")
	client := openai.NewClientWithConfig(config)

	ctx := context.Background()
	req := openai.CompletionRequest{Stream: true}
	_, err := client.CreateCompletion(ctx, req)
	if !errors.Is(err, openai.ErrCompletionStreamNotSupported) {
		t.Fatalf("CreateCompletion didn't return ErrCompletionStreamNotSupported")
	}
}

// TestCompletions Tests the completions endpoint of the API using the mocked server.
func TestCompletions(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/completions", handleCompletionEndpoint)
	req := openai.CompletionRequest{
		MaxTokens: 5,
		Model:     "ada",
		Prompt:    "Lorem ipsum",
	}
	_, err := client.CreateCompletion(context.Background(), req)
	checks.NoError(t, err, "CreateCompletion error")
}

// handleCompletionEndpoint Handles the completion endpoint by the test server.
func handleCompletionEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// completions only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	var completionReq openai.CompletionRequest
	if completionReq, err = getCompletionBody(r); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}
	prompt, ok := completionReq.Prompt.(string)
	if !ok {
		http.Error(w, "expected a string prompt", http.StatusBadRequest)
		return
	}
	res := openai.CompletionResponse{
		ID:      strconv.Itoa(int(time.Now().Unix())),
		Object:  "test-object",
		Created: time.Now().Unix(),
		// would be nice to validate Model during testing, but
		// this may not be possible with how much upkeep
		// would be required / wouldn't make much sense
		Model: completionReq.Model,
	}
	// create completions
	n := completionReq.N
	if n == 0 {
		n = 1
	}
	for i := 0; i < n; i++ {
		// generate a random string of length completionReq.Length
		completionStr := strings.Repeat("a", completionReq.MaxTokens)
		if completionReq.Echo {
			completionStr = prompt + completionStr
		}
		res.Choices = append(res.Choices, openai.CompletionChoice{
			Text:  completionStr,
			Index: i,
		})
	}
	inputTokens := numTokens(prompt) * n
	completionTokens := completionReq.MaxTokens * n
	res.Usage = openai.Usage{
		PromptTokens:     inputTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      inputTokens + completionTokens,
	}
	resBytes, _ = json.Marshal(res)
	fmt.Fprintln(w, string(resBytes))
}

// getCompletionBody Returns the body of the request to create a completion.
func getCompletionBody(r *http.Request) (openai.CompletionRequest, error) {
	completion := openai.CompletionRequest{}
	// read the request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return openai.CompletionRequest{}, err
	}
	err = json.Unmarshal(reqBody, &completion)
	if err != nil {
		return openai.CompletionRequest{}, err
	}
	return completion, nil
}

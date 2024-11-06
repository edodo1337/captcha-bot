package clients

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiClient struct {
	client     *genai.Client
	model      *genai.GenerativeModel
	promptWrap string
}

func NewGeminiClient(ctx context.Context, token string, modelType string, promptWrap string) *GeminiClient {
	geminiClient, err := genai.NewClient(ctx, option.WithAPIKey(token))
	if err != nil {
		log.Println(err)
		return nil
	}
	model := geminiClient.GenerativeModel(modelType)
	return &GeminiClient{
		client:     geminiClient,
		model:      model,
		promptWrap: promptWrap,
	}
}

func (c *GeminiClient) Shutdown() error {
	return c.client.Close()
}

func (c *GeminiClient) IsSpam(ctx context.Context, text string) (bool, error) {
	var b strings.Builder
	b.WriteString(c.promptWrap)
	b.WriteString(text)
	promptText := b.String()

	resp, err := c.model.StartChat().SendMessage(ctx, genai.Text(promptText))
	if err != nil {
		log.Println(err, resp)
		return false, err
	}

	if len(resp.Candidates) < 1 {
		return false, errors.New("empty response candidates")
	}
	candidate := resp.Candidates[0]

	if len(candidate.Content.Parts) < 1 {
		return false, errors.New("empty candidate parts")
	}
	responseText := fmt.Sprint(candidate.Content.Parts[0])
	switch responseText {
	case "1":
		return true, nil
	case "0":
		return false, nil
	default:
		return false, fmt.Errorf("unexpected response: %s", responseText)
	}
}

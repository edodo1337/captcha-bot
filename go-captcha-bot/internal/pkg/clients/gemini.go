package clients

import (
	"captcha-bot/internal/pkg/conf"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiClient struct {
	client *genai.Client
	config *conf.Config
	model  *genai.GenerativeModel
}

func NewGeminiClient(ctx context.Context, config *conf.Config) *GeminiClient {
	geminiClient, err := genai.NewClient(ctx, option.WithAPIKey(config.Bot.GeminiApiToken))
	if err != nil {
		log.Fatal(err)
		return nil
	}
	model := geminiClient.GenerativeModel(config.Bot.GeminiModel)

	return &GeminiClient{
		config: config,
		client: geminiClient,
		model:  model,
	}
}

func (c *GeminiClient) Shutdown() error {
	return c.client.Close()
}

func (c *GeminiClient) IsSpam(ctx context.Context, text string) (bool, error) {
	var b strings.Builder
	b.WriteString(c.config.Bot.PromptWrap)
	b.WriteString(text)
	promptText := b.String()

	resp, err := c.model.StartChat().SendMessage(ctx, genai.Text(promptText))
	if err != nil {
		log.Fatal(err)
	}

	if len(resp.Candidates) < 1 {
		return false, errors.New("Empty response candidates")
	}
	candidate := resp.Candidates[0]

	if len(candidate.Content.Parts) < 1 {
		return false, errors.New("Empty candidate parts")
	}
	responseText := fmt.Sprint(candidate.Content.Parts[0])
	switch responseText {
	case "1":
		return true, nil
	case "0":
		return false, nil
	default:
		return false, errors.New("Unexpected response")
	}
}

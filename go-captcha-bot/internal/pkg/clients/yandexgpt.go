package clients

import (
	"context"
	"fmt"
	"strings"

	"github.com/sheeiavellie/go-yandexgpt"
)

type YandexGPTClient struct {
	client     *yandexgpt.YandexGPTClient
	promptWrap string
	modelUri   string
}

func NewYandexGPTClient(ctx context.Context, promptWrap string, catalogID string, token string) *YandexGPTClient {
	client := yandexgpt.NewYandexGPTClientWithAPIKey(token)
	modelUri := yandexgpt.MakeModelURI(catalogID, yandexgpt.YandexGPTModelLite)
	return &YandexGPTClient{
		client:     client,
		promptWrap: promptWrap,
		modelUri:   modelUri,
	}
}

func (c *YandexGPTClient) Shutdown() error {
	return nil
}

func (c *YandexGPTClient) IsSpam(ctx context.Context, text string) (bool, error) {
	var b strings.Builder
	b.WriteString(c.promptWrap)
	b.WriteString(text)
	promptText := b.String()

	request := yandexgpt.YandexGPTRequest{
		ModelURI: c.modelUri,
		CompletionOptions: yandexgpt.YandexGPTCompletionOptions{
			Stream:      false,
			Temperature: 0.7,
			MaxTokens:   2000,
		},
		Messages: []yandexgpt.YandexGPTMessage{
			{
				Role: yandexgpt.YandexGPTMessageRoleSystem,
				Text: promptText,
			},
		},
	}

	response, err := c.client.CreateRequest(ctx, request)
	if err != nil {
		return false, err
	}

	if len(response.Result.Alternatives) == 0 {
		return false, fmt.Errorf("empty response candidates")
	}

	responseText := response.Result.Alternatives[0].Message.Text
	switch responseText {
	case "1":
		return true, nil
	case "0":
		return false, nil
	default:
		return false, fmt.Errorf("unexpected response: %s", responseText)
	}
}

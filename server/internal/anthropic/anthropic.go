package anthropic

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"byakugan/internal/store"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type Client struct {
	api anthropic.Client
}

func New(apiKey string) *Client {
	return &Client{api: anthropic.NewClient(option.WithAPIKey(apiKey))}
}

//go:embed prompts/framing_v1.txt
var systemPrompt string

func (c *Client) Frame(ctx context.Context, question string, hits []store.Hit) (string, bool, error) {

	var sb strings.Builder
	fmt.Fprintf(&sb, "Question: %s\n\nRetrieved sections:\n", question)
	for _, h := range hits {
		fmt.Fprintf(&sb, "[s%s %s]\n%s\n\n", h.Section, h.Heading, h.Text)
	}
	userContent := sb.String()

	msg, err := c.api.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 1024,
		System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userContent)),
		},
	})
	truncated := msg.StopReason != anthropic.StopReasonEndTurn
	if err != nil {
		return "", false, fmt.Errorf("framing failed: %w", err)
	}

	for _, block := range msg.Content {
		if t, ok := block.AsAny().(anthropic.TextBlock); ok {
			return t.Text, truncated, nil
		}
	}

	return "", false, fmt.Errorf("no text block in res")
}

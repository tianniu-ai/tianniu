package context

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/tianniu-ai/tianniu/pkg/shared"
	"github.com/tianniu-ai/tianniu/pkg/storage"
)

type OffloadPolicy struct {
	// Storage is used to save offloaded long text content.
	Storage storage.Storage
	// UsageThreshold triggers offloading when context usage exceeds this value.
	UsageThreshold float64
	// KeepRecentMessages skips the last N messages to avoid affecting the latest conversation.
	KeepRecentMessages int
	// PreviewCharLimit is the number of characters to keep in context after offloading.
	PreviewCharLimit int
}

func NewOffloadPolicy(storage storage.Storage, usageThreshold float64, keepRecentMessages, previewCharLimit int) *OffloadPolicy {
	return &OffloadPolicy{
		Storage:            storage,
		UsageThreshold:     usageThreshold,
		KeepRecentMessages: keepRecentMessages,
		PreviewCharLimit:   previewCharLimit,
	}
}

func (p *OffloadPolicy) Name() string {
	return "offload"
}

func (p *OffloadPolicy) makeStorageKey(conversationId string, offloadIndex int) string {
	return fmt.Sprintf("/offload/%s/%s_%d", conversationId, time.Now().Format("20060102_150405"), offloadIndex)
}

func (p *OffloadPolicy) Apply(ctx context.Context, engine *Engine) (PolicyResult, error) {
	if len(engine.messages) <= p.KeepRecentMessages {
		return PolicyResult{
			Messages:      engine.messages,
			ContextTokens: engine.contextTokens,
		}, nil
	}

	conversationId, ok := ctx.Value("conversationId").(string)
	if !ok {
		return PolicyResult{}, errors.New("conversationId not found in context")
	}

	// Copy message list to avoid modifying original data
	messages := make([]messageWrap, len(engine.messages))
	copy(messages, engine.messages)
	contextTokens := engine.contextTokens

	offloadCount := len(messages) - p.KeepRecentMessages

	for i := 0; i < offloadCount; i++ {
		// Only offload tool type messages
		if shared.GetRoleName(messages[i].Message) != "tool" {
			continue
		}

		contentAny := messages[i].Message.GetContent().AsAny()
		contentStr, ok := contentAny.(*string)
		if !ok {
			continue
		}
		// No need to offload
		if len(*contentStr) <= p.PreviewCharLimit {
			continue
		}

		// Calculate token count of original message
		oldTokens := messages[i].Tokens

		key := p.makeStorageKey(conversationId, i)
		if err := p.Storage.Store(ctx, key, *contentStr); err != nil {
			log.Printf("failed to store offload message: %v", err)
			continue
		}

		// Construct offloaded message content
		abstract := (*contentStr)[0:p.PreviewCharLimit]
		var b strings.Builder
		b.WriteString(abstract)
		b.WriteString("...")
		b.WriteString(fmt.Sprintf("（更多内容已卸载，如需查看全文请使用 load_storage(key=\"%s\") 工具）\n", key))
		newContent := b.String()

		// Modify message in original message chain
		newMessage := openai.ToolMessage(newContent, *engine.messages[i].Message.GetToolCallID())

		// Calculate token count for new message and update total
		newTokens := CountTokens(newMessage)
		messages[i] = messageWrap{Message: newMessage, Tokens: newTokens}
		contextTokens -= oldTokens - newTokens
	}

	return PolicyResult{
		Messages:      messages,
		ContextTokens: contextTokens,
	}, nil
}

func (p *OffloadPolicy) ShouldApply(ctx context.Context, engine *Engine) bool {
	return engine.GetContextUsage() > p.UsageThreshold
}

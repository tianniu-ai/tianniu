package context

import (
	"context"

	"github.com/openai/openai-go/v3"
	log "github.com/sirupsen/logrus"
	"github.com/tianniu-ai/tianniu/pkg/shared"
)

type SummaryPolicy struct {
	// KeepRecentMessages indicates to skip the last N messages to avoid summarizing the latest conversation.
	KeepRecentMessages int
	// SummaryBatchSize indicates the maximum number of messages sent to the summarizer at one time.
	SummaryBatchSize int
	// UsageThreshold triggers summarization when context usage exceeds this value.
	UsageThreshold float64
	// Summarizer is responsible for generating summaries.
	Summarizer Summarizer
}

func (p *SummaryPolicy) Name() string {
	return "summarize"
}

func NewSummaryPolicy(summarizer Summarizer, keepRecentMessages, summaryBatchSize int, usageThreshold float64) *SummaryPolicy {
	return &SummaryPolicy{
		KeepRecentMessages: keepRecentMessages,
		Summarizer:         summarizer,
		SummaryBatchSize:   summaryBatchSize,
		UsageThreshold:     usageThreshold,
	}
}

func (p *SummaryPolicy) ShouldApply(ctx context.Context, engine *Engine) bool {
	return engine.GetContextUsage() > p.UsageThreshold
}

func (p *SummaryPolicy) Apply(ctx context.Context, engine *Engine) (PolicyResult, error) {
	if len(engine.messages) <= p.KeepRecentMessages {
		return PolicyResult{
			Messages:      engine.messages,
			ContextTokens: engine.contextTokens,
		}, nil
	}

	summarizeUntilIndex := len(engine.messages) - p.KeepRecentMessages
	inputTokenLimit := p.Summarizer.GetSummaryInputTokenLimit()

	accumulatedSummary := ""

	// Calculate total tokens of messages to be replaced
	removedTokens := 0
	for i := 0; i < summarizeUntilIndex; i++ {
		removedTokens += engine.messages[i].Tokens
	}

	batchStart := 0

	for batchStart < summarizeUntilIndex {
		batchMessages := make([]shared.OpenAIMessage, 0)
		batchTokens := 0

		for i := batchStart; i < summarizeUntilIndex; i++ {
			// Calculate token count for current message
			msgTokens := engine.messages[i].Tokens

			// Stop adding if threshold exceeded and there are already messages
			if batchTokens+msgTokens > inputTokenLimit && len(batchMessages) > 0 {
				break
			}

			batchMessages = append(batchMessages, engine.messages[i].Message)
			batchTokens += msgTokens

			// Reached batch size, stop adding
			if len(batchMessages) >= p.SummaryBatchSize {
				break
			}
		}

		if len(batchMessages) == 0 {
			break
		}

		batchSummary, err := p.Summarizer.Summarize(ctx, accumulatedSummary, batchMessages)
		if err != nil {
			log.Errorf("Summarize: %v", err)
			return PolicyResult{}, err
		}

		accumulatedSummary = batchSummary
		batchStart += len(batchMessages)
	}

	if len(accumulatedSummary) == 0 {
		log.Infof("no summary generated")
		return PolicyResult{
			Messages:      engine.messages,
			ContextTokens: engine.contextTokens,
		}, nil
	}

	// Build new message list
	messages := make([]messageWrap, 0, len(engine.messages))

	summaryMessage := openai.UserMessage(accumulatedSummary)
	newTokens := CountTokens(summaryMessage)

	messages = append(messages, messageWrap{Message: summaryMessage, Tokens: newTokens})
	messages = append(messages, engine.messages[summarizeUntilIndex:]...)

	// Return new message list and token count
	return PolicyResult{
		Messages:      messages,
		ContextTokens: engine.contextTokens - removedTokens + newTokens,
	}, nil
}

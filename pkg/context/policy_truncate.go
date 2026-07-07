package context

import "context"

type TruncatePolicy struct {
	// KeepRecentMessages indicates the minimum number of recent messages to keep.
	KeepRecentMessages int
	// UsageThreshold triggers truncation when context usage exceeds this value.
	UsageThreshold float64
}

func NewTruncatePolicy(keepRecentMessages int, usageThreshold float64) *TruncatePolicy {
	return &TruncatePolicy{
		KeepRecentMessages: keepRecentMessages,
		UsageThreshold:     usageThreshold,
	}
}

func (p *TruncatePolicy) Name() string {
	return "truncate"
}

func (p *TruncatePolicy) Apply(ctx context.Context, engine *Engine) (PolicyResult, error) {
	if len(engine.messages) <= p.KeepRecentMessages {
		return PolicyResult{
			Messages:      engine.messages,
			ContextTokens: engine.contextTokens,
		}, nil
	}

	// Prepare to remove the first toRemove messages
	toRemove := len(engine.messages) - p.KeepRecentMessages

	// Find the last User message in 0 ~ toRemove-1, keep messages after this User, truncate all history before
	removeIdx := toRemove - 1
	for i := toRemove - 1; i >= 0; i-- {
		if engine.messages[i].Message.OfUser != nil {
			removeIdx = i
			break
		}
	}

	// If no user message found or removeIdx is 0, do not remove any messages
	// This ensures we never delete all messages
	if removeIdx <= 0 {
		return PolicyResult{
			Messages:      engine.messages,
			ContextTokens: engine.contextTokens,
		}, nil
	}

	removedTokens := 0
	for i := 0; i < removeIdx; i++ {
		removedTokens += engine.messages[i].Tokens
	}

	return PolicyResult{
		Messages:      engine.messages[removeIdx:],
		ContextTokens: engine.contextTokens - removedTokens,
	}, nil
}

func (p *TruncatePolicy) ShouldApply(ctx context.Context, engine *Engine) bool {
	return engine.GetContextUsage() > p.UsageThreshold
}

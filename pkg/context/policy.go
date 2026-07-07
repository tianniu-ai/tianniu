package context

import (
	"context"
)

// PolicyResult represents the execution result of a policy
type PolicyResult struct {
	Messages      []messageWrap // New message list
	ContextTokens int           // New context token count
}

type Policy interface {
	Name() string
	ShouldApply(ctx context.Context, engine *Engine) bool
	// Apply is a pure function that reads engine state and returns new state without modifying internal variables
	Apply(ctx context.Context, engine *Engine) (PolicyResult, error)
}

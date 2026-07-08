package context

import (
	"encoding/json"

	"github.com/tianniu-ai/tianniu/pkg/model"

	"github.com/tianniu-ai/tianniu/pkg/shared"
)

// buildHistory traces the path from parent_message_id up to the root,
// concatenating the rounds of each message along the path into LLM history.
// allMsgs contains all messages in the conversation; parentMessageID is the parent message ID for this request.
func buildHistory(allMsgs []*model.ChatMessage, parentMessageID string) []shared.OpenAIMessage {
	if parentMessageID == "" {
		return nil
	}

	// Build id -> message index
	index := make(map[string]*model.ChatMessage, len(allMsgs))
	for i := range allMsgs {
		index[allMsgs[i].ID] = allMsgs[i]
	}

	// Trace from parentMessageID to root, collecting the path (order: leaf -> root)
	path := make([]*model.ChatMessage, 0)
	cur := parentMessageID
	for cur != "" {
		msg, ok := index[cur]
		if !ok {
			break
		}
		path = append(path, msg)
		cur = msg.ParentMessageID
	}

	// Reverse to root -> parent order
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	// Concatenate rounds from each message
	history := make([]shared.OpenAIMessage, 0)
	for _, msg := range path {
		if msg.Rounds == "" {
			continue
		}
		var rounds []shared.OpenAIMessage
		if err := json.Unmarshal([]byte(msg.Rounds), &rounds); err != nil {
			continue
		}
		history = append(history, rounds...)
	}
	return history
}

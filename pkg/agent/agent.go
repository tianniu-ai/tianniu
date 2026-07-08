package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
	ctxengine "github.com/tianniu-ai/tianniu/pkg/agent/context"
	"github.com/tianniu-ai/tianniu/pkg/agent/llm"
	"github.com/tianniu-ai/tianniu/pkg/agent/mcp"
	"github.com/tianniu-ai/tianniu/pkg/agent/tool"
	"github.com/tianniu-ai/tianniu/pkg/shared"
)

type Agent struct {
	model         string
	client        openai.Client
	nativeTools   map[tool.AgentTool]tool.Tool
	systemPrompt  string
	mcpClients    map[string]*mcp.Client
	contextEngine *ctxengine.Engine
}

func NewAgent(modelConf shared.ModelConfig,
	systemPrompt string,
	tools []tool.Tool,
	mcpClients []*mcp.Client,
	contextEngine *ctxengine.Engine) *Agent {
	a := &Agent{
		model:         modelConf.Model,
		client:        llm.NewLLMClient(modelConf),
		nativeTools:   make(map[tool.AgentTool]tool.Tool),
		systemPrompt:  systemPrompt,
		mcpClients:    make(map[string]*mcp.Client),
		contextEngine: contextEngine,
	}
	for _, t := range tools {
		a.nativeTools[t.ToolName()] = t
	}
	for _, mcpClient := range mcpClients {
		a.mcpClients[mcpClient.Name()] = mcpClient
	}
	a.contextEngine.Init(systemPrompt, ctxengine.TokenBudget{ContextWindow: modelConf.ContextWindow})
	return a
}

func (a *Agent) Model() string {
	return a.model
}

func (a *Agent) findTool(toolName string) (tool.Tool, bool) {
	t, ok := a.nativeTools[toolName]
	if ok {
		return t, true
	}
	for _, mcpClient := range a.mcpClients {
		for _, t := range mcpClient.GetTools() {
			if t.ToolName() != toolName {
				continue
			}
			return t, true
		}
	}
	return t, false
}

func (a *Agent) buildTools() []openai.ChatCompletionToolUnionParam {
	tools := make([]openai.ChatCompletionToolUnionParam, 0, len(a.nativeTools))
	for _, t := range a.nativeTools {
		tools = append(tools, t.Info())
	}
	for _, mcpClient := range a.mcpClients {
		for _, t := range mcpClient.GetTools() {
			tools = append(tools, t.Info())
		}
	}
	return tools
}

// executeTool executes a single tool call, returning the tool result and error.
// Returns an error if the tool is not found; if Execute fails, the error message is returned as result.
func (a *Agent) executeTool(ctx context.Context, toolCall openai.ChatCompletionMessageToolCallUnion) (string, error) {
	t, ok := a.findTool(toolCall.Function.Name)
	if !ok {
		return "", fmt.Errorf("tool not found: %s", toolCall.Function.Name)
	}
	return t.Execute(ctx, toolCall.Function.Arguments)
}

// RunResult holds the result of one agent run
type RunResult struct {
	Response string
	Rounds   []shared.OpenAIMessage
	Usage    openai.CompletionUsage
}

// RunStreaming executes the agent loop, streaming output via eventCh, and returns RunResult when done.
// history is the deserialized message list from all previous ChatMessage.Rounds in this conversation.
func (a *Agent) RunStreaming(ctx context.Context, query string, eventCh chan<- StreamEvent) (RunResult, error) {

	draft := a.contextEngine.StartTurn(openai.UserMessage(query))
	defer a.contextEngine.AbortTurn(draft)

	messages := a.contextEngine.BuildRequestMessages()
	messages = append(messages, draft.NewMessages...)
	var usage openai.CompletionUsage

	var finalResponse string

	for {
		params := openai.ChatCompletionNewParams{
			Model:         a.model,
			Messages:      messages,
			Tools:         a.buildTools(),
			StreamOptions: openai.ChatCompletionStreamOptionsParam{IncludeUsage: openai.Bool(true)},
		}

		stream := a.client.Chat.Completions.NewStreaming(ctx, params)
		acc := openai.ChatCompletionAccumulator{}

		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			if len(chunk.Choices) > 0 {
				deltaRaw := chunk.Choices[0].Delta
				delta := deltaWithReasoning{}
				_ = json.Unmarshal([]byte(deltaRaw.RawJSON()), &delta)

				if delta.ReasoningContent != "" {
					eventCh <- StreamEvent{Event: EventReasoning, ReasoningContent: delta.ReasoningContent}
				}
				if delta.Content != "" {
					eventCh <- StreamEvent{Event: EventContent, Content: delta.Content}
				}
			}
		}
		if err := stream.Err(); err != nil {
			eventCh <- StreamEvent{Event: EventError, Content: err.Error()}
			return RunResult{}, err
		}
		if len(acc.Choices) == 0 {
			break
		}

		usage = acc.Usage
		message := acc.Choices[0].Message
		assistantMsg := message.ToParam()
		messages = append(messages, assistantMsg)

		// No tool calls, end loop
		if len(message.ToolCalls) == 0 {
			finalResponse = message.Content
			break
		}

		// Execute tool calls
		for _, toolCall := range message.ToolCalls {
			eventCh <- StreamEvent{Event: EventToolCall, ToolCall: toolCall.Function.Name, ToolArguments: toolCall.Function.Arguments}

			toolResult, err := a.executeTool(ctx, toolCall)
			if err != nil {
				toolResult = err.Error()
				eventCh <- StreamEvent{Event: EventError, Content: toolResult}
			}
			eventCh <- StreamEvent{Event: EventToolResult, ToolCall: toolCall.Function.Name, ToolResult: toolResult}

			toolMsg := openai.ToolMessage(toolResult, toolCall.ID)
			messages = append(messages, toolMsg)

			messages = append(messages, toolMsg)
			draft.NewMessages = append(draft.NewMessages, toolMsg)
		}

		// Check if context is canceled
		select {
		case <-ctx.Done():
			return RunResult{Response: finalResponse}, ctx.Err()
		default:
		}
	}

	err := a.contextEngine.CommitTurn(ctx, draft, ctxengine.Usage{PromptTokens: int(usage.TotalTokens)})
	if err != nil {
		return RunResult{}, err
	}

	return RunResult{
		Response: finalResponse,
		Rounds:   messages,
		Usage:    usage,
	}, nil
}

type deltaWithReasoning struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
}

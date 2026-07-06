package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"

	"github.com/liyue201/tian-niu/pkg/agent/tool"
	"github.com/liyue201/tian-niu/pkg/shared"
)

const SystemPrompt = `# 天牛

You are 天牛, a professional knowledge Q&A assistant.

## Guidelines
- Answers may draw upon the provided knowledge base. If no relevant materials are available, you may respond based on your existing knowledge.
- For complex questions, conduct step-by-step reasoning: break down requirements, filter documents, and verify information before reaching conclusions. Separate reasoning processes from the final response.
- When faced with vague or incomplete inquiries, proactively guide users to supply critical conditions; do not cobble together invalid answers.
- Present comparison questions in structured tables, and provide scenario-based selection recommendations at the end.
- Keep answers concise and well-organized with clear paragraphs and bullet points. Use precise professional terminology and avoid irrelevant chatter.
- Wrap all code snippets with Markdown syntax highlighting blocks.
`

type Agent struct {
	model        string
	client       openai.Client
	nativeTools  map[tool.AgentTool]tool.Tool
	systemPrompt string
	mcpClients   map[string]*McpClient
}

func NewAgent(modelConf shared.ModelConfig, systemPrompt string, tools []tool.Tool,
	mcpClients []*McpClient) *Agent {
	a := &Agent{
		model:        modelConf.Model,
		client:       shared.NewLLMClient(modelConf),
		nativeTools:  make(map[tool.AgentTool]tool.Tool),
		systemPrompt: systemPrompt,
		mcpClients:   make(map[string]*McpClient),
	}
	for _, t := range tools {
		a.nativeTools[t.ToolName()] = t
	}
	for _, mcpClient := range mcpClients {
		a.mcpClients[mcpClient.Name()] = mcpClient
	}
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
func (a *Agent) RunStreaming(ctx context.Context, history []openai.ChatCompletionMessageParamUnion, query string, eventCh chan<- StreamEvent) (RunResult, error) {
	// Build messages for this round: system + history + current user message
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(history)+2)
	messages = append(messages, openai.SystemMessage(a.systemPrompt))
	messages = append(messages, history...)
	messages = append(messages, openai.UserMessage(query))

	// roundMessages tracks new messages from this round (user + assistant + tool, excluding system and history)
	roundMessages := []shared.OpenAIMessage{openai.UserMessage(query)}

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
		roundMessages = append(roundMessages, assistantMsg)

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
			roundMessages = append(roundMessages, toolMsg)
		}

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return RunResult{Response: finalResponse}, ctx.Err()
		default:
		}
	}

	return RunResult{
		Response: finalResponse,
		Rounds:   roundMessages,
		Usage:    usage,
	}, nil
}

type deltaWithReasoning struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
}

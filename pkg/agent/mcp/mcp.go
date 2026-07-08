package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openai/openai-go/v3"
	shared2 "github.com/openai/openai-go/v3/shared"
	"github.com/tianniu-ai/tianniu/pkg/agent/tool"
	"github.com/tianniu-ai/tianniu/pkg/shared"
)

type Client struct {
	name         string
	client       *mcp.Client
	serverConfig ServerConfig

	session *mcp.ClientSession
	tools   []tool.Tool
}

func initRunningVars() map[string]string {
	runningVars := map[string]string{
		"${workspaceFolder}": shared.GetWorkspaceDir(),
	}
	return runningVars
}

func NewMcpToolProvider(name string, server ServerConfig) *Client {

	return &Client{
		name: name,
		client: mcp.NewClient(&mcp.Implementation{
			Name:    "tianniu-mcp-client",
			Title:   "TianNiu",
			Version: "v1.0.0",
		}, nil),
		serverConfig: server.ReplacePlaceholders(initRunningVars()),
		tools:        make([]tool.Tool, 0),
	}
}

func (e *Client) Name() string {
	return e.name
}

func (e *Client) connect(ctx context.Context) error {
	if e.session != nil && e.session.Ping(ctx, &mcp.PingParams{}) == nil {
		return nil
	}
	var err error
	if e.serverConfig.IsStdio() {
		cmd := exec.Command(e.serverConfig.Command, e.serverConfig.Args...)
		for k, v := range e.serverConfig.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
		e.session, err = e.client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	} else {
		e.session, err = e.client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: e.serverConfig.Url}, nil)
	}
	if err != nil {
		return err
	}

	return nil
}

func (e *Client) RefreshTools(ctx context.Context) error {
	if err := e.connect(ctx); err != nil {
		return err
	}

	mcpToolResult, err := e.session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return err
	}

	e.tools = make([]tool.Tool, 0)
	for _, mcpTool := range mcpToolResult.Tools {
		agentTool := &Tool{
			client:   e,
			toolName: mcpTool.Name,
			session:  e.session,
			mcpTool:  mcpTool,
		}

		e.tools = append(e.tools, agentTool)
	}
	return nil
}

func (e *Client) GetTools() []tool.Tool {
	return e.tools
}

func (e *Client) callTool(ctx context.Context, toolName string, argumentsInJSON string) (string, error) {
	if err := e.connect(ctx); err != nil {
		return "", err
	}
	mcpResult, err := e.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: json.RawMessage(argumentsInJSON),
	})
	if err != nil {
		log.Printf("failed to call tool: %v", err)
		return "", err
	}

	var builder strings.Builder
	for _, content := range mcpResult.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			builder.WriteString(textContent.Text)
		}
	}
	return builder.String(), nil
}

// Tool implements the tool.Tool interface
type Tool struct {
	toolName string // name sent to the MCP server; differs from the name exposed to the model
	client   *Client
	session  *mcp.ClientSession
	mcpTool  *mcp.Tool
}

// ToolName returns the name exposed to the model; differs from the name sent to the MCP server
func (t *Tool) ToolName() string {
	return fmt.Sprintf("tianniu_mcp__%s__%s", t.client.Name(), t.toolName)
}

func (t *Tool) Info() openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionFunctionTool(shared2.FunctionDefinitionParam{
		Description: openai.String(t.mcpTool.Description),
		Name:        t.ToolName(),
		Parameters:  t.mcpTool.InputSchema.(map[string]any),
	})
}

func (t *Tool) Execute(ctx context.Context, argumentsInJSON string) (string, error) {
	return t.client.callTool(ctx, t.toolName, argumentsInJSON)
}

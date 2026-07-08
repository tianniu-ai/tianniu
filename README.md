# 天牛

A lightweight AI chat agent built with Go, featuring streaming message processing, tool calls, MCP integration, and multi-threaded conversations.

## Features

- **Smart Conversations**: Fluent AI model interaction with streaming responses
- **Multi-thread Management**: Create, rename, and delete conversation threads
- **Streaming Messages**: Real-time message delivery via Server-Sent Events (SSE)
- **Tool Calls**: AI can invoke external tools to fetch information
- **MCP Integration**: Connect to any MCP-compatible tool server (stdio / HTTP) to extend agent capabilities
- **Reasoning Panel**: Display the AI's thinking process (DeepSeek-R1, QwQ, etc.)
- **Bash Tool**: Execute shell commands with built-in security restrictions (dangerous pattern blocking, timeout, output limits, env filtering)
- **Memory System**: Multi-level memory management (global + workspace) for context retention
- **Markdown Rendering**: Full markdown support (GFM) for AI responses and tool results
- **JWT Authentication**: User registration, login, and token-based access control
- **Context Management**: Automatic message summarization and content offloading to manage context window

## Tech Stack

### Backend
- Go 1.26.4 + Gin
- OpenAI Go SDK v3
- MCP Go SDK v1
- GORM + SQLite
- Redis for memory storage
- JWT authentication (golang-jwt/v5)

## Quick Start

### Local Development

1. **Configure the backend**

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml` with your LLM provider and tool settings:


2. **(Optional) Configure MCP servers**

Edit `mcp-server.json` to connect external tool servers:

3. **Start the backend**

```bash
go run ./tianniu/main.go
```
The server runs on `http://localhost:8080`.

## Frontend Integration
Refer to the workspace repository: https://github.com/tianniu-ai/tianniu-workspace

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_PATH` | SQLite database file path | `test.db` |
| `LEVELDB_PATH` | LevelDB storage path for memory and offloaded content | `leveldb_data` |
| `JWT_SECRET` | Secret key for JWT token signing (≥16 bytes) | `tian-niu-dev-secret-change-in-production` |
| `GIN_MODE` | Gin run mode (`debug`/`release`) | `debug` |

## Configuration

### Context Management Policies

The agent automatically manages context window usage through two policies:

#### Summary Policy

When context usage exceeds the threshold, older messages are summarized to reduce token count:

- **KeepRecentMessages**: Number of recent messages to skip (avoid summarizing latest conversation)
- **SummaryBatchSize**: Maximum messages to summarize at one time
- **UsageThreshold**: Context usage percentage that triggers summarization

#### Offload Policy

When context usage exceeds the threshold, large tool response content is offloaded to storage:

- **KeepRecentMessages**: Number of recent messages to skip
- **PreviewCharLimit**: Number of characters to keep in context as preview
- **UsageThreshold**: Context usage percentage that triggers offloading

When content is offloaded, the agent replaces it with a preview and provides a `load_storage(key="...")` call suggestion to retrieve the full content when needed.

#### Truncate Policy

When context usage exceeds the threshold, older messages are truncated to reduce token count:

- **KeepRecentMessages**: Number of recent messages to skip
- **UsageThreshold**: Context usage percentage that triggers truncation

### MCP Server Configuration

MCP (Model Context Protocol) allows the agent to use external tool servers. Configure in `mcp-server.json`:

**Stdio transport** (local process):

```json
{
  "filesystem": {
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/dir"]
  }
}
```

**HTTP transport** (remote server):

```json
{
  "remote-api": {
    "type": "http",
    "url": "http://localhost:3001/mcp",
    "headers": {
      "Authorization": "Bearer token"
    }
  }
}
```

MCP tools are automatically discovered and registered as agent tools at startup. Tool names are prefixed as `babyagent_mcp__<server>__<tool>` to avoid conflicts.

## Supported Models

- OpenAI: gpt-4o, gpt-4o-mini, gpt-4-turbo, gpt-5.2
- DeepSeek: deepseek-chat, deepseek-reasoner (with reasoning output)
- Zhipu AI: GLM-5.2, GLM-4, GLM-4.6V
- Qwen: QwQ, Qwen3
- Any model compatible with the OpenAI API format

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/user/register` | No | Register a new user |
| POST | `/api/user/login` | No | Login and get JWT token |
| POST | `/api/conversation` | Yes | Create a conversation |
| GET | `/api/conversation` | Yes | List user's conversations |
| PATCH | `/api/conversation/:id` | Yes | Rename a conversation |
| DELETE | `/api/conversation/:id` | Yes | Delete a conversation |
| POST | `/api/conversation/:id/message` | Yes | Send a message (SSE stream) |
| GET | `/api/conversation/:id/message` | Yes | List conversation messages |

## Preview

![Effect](./doc/tianniu.png)

## License

MIT License
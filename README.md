# 天牛

A lightweight AI chat agent built with Go and React, featuring streaming message processing, tool calls, MCP integration, and multi-threaded conversations.

## Features

- **Smart Conversations**: Fluent AI model interaction with streaming responses
- **Multi-thread Management**: Create, rename, and delete conversation threads
- **Streaming Messages**: Real-time message delivery via Server-Sent Events (SSE)
- **Tool Calls**: AI can invoke external tools to fetch information
- **MCP Integration**: Connect to any MCP-compatible tool server (stdio / HTTP) to extend agent capabilities
- **Reasoning Panel**: Display the AI's thinking process (DeepSeek-R1, QwQ, etc.)
- **Bash Tool**: Execute shell commands with built-in security restrictions (dangerous pattern blocking, timeout, output limits, env filtering)
- **Markdown Rendering**: Full markdown support (GFM) for AI responses and tool results
- **JWT Authentication**: User registration, login, and token-based access control

## Tech Stack

### Frontend
- React 19 + TypeScript
- Vite 8
- Tailwind CSS 4.0
- @assistant-ui/react, @radix-ui/react
- react-markdown + remark-gfm
- lucide-react

### Backend
- Go 1.25 + Gin
- OpenAI Go SDK v3
- MCP Go SDK v1
- GORM + SQLite
- JWT authentication (golang-jwt/v5)

## Architecture

```
tianniu/main.go              — Entry point
pkg/
  agent/                     — Agent core (LLM loop + tool dispatch)
    tool/                    — Native tool definitions (Bash, etc.)
    mcp.go                   — MCP client integration
    stream.go                — Streaming event types
  server/                    — HTTP layer (Gin routes + SSE)
  service/                   — Business logic
  repository/                — Data access (GORM + SQLite)
  model/                     — Data models
  vo/                        — View objects (request/response/SSE)
  auth/                      — JWT authentication
  shared/                    — Shared utilities & config
```

## Quick Start

### Local Development

1. **Configure the backend**

```bash
cp config.example.json config.json
```

Edit `config.json` with your LLM provider and tool settings:

```json
{
  "llm_providers": {
    "front_model": {
      "base_url": "https://api.openai.com/v1",
      "model": "gpt-4o",
      "api_key": "your-api-key",
      "context_window": 200000
    }
  },
  "bash_tool": {
    "timeout_seconds": 30,
    "max_output_kb": 64,
    "work_dir": "",
    "disabled": false,
    "allow_dangerous": false
  }
}
```

2. **(Optional) Configure MCP servers**

```bash
cp mcp-server.example.json mcp-server.json
```

Edit `mcp-server.json` to connect external tool servers:

```json
{
  "filesystem": {
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/dir"]
  },
  "remote-api": {
    "type": "http",
    "url": "http://localhost:3001/mcp"
  }
}
```

3. **Start the backend**

```bash
go run ./tianniu/main.go
```

The server runs on `http://localhost:8080`.

4. **Start the frontend**

```bash
cd frontend
npm install
npm run dev
```

Visit `http://localhost:5173`. API requests to `/api` are proxied to the backend automatically via Vite.

### Docker Deployment

1. **Prepare config**

```bash
cp config.example.json config.json
# Edit config.json with your API keys
```

2. **Build and run**

```bash
docker compose up -d --build
```

3. **Access the app**

- Frontend: http://localhost:80
- Backend API: http://localhost:8080

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_PATH` | SQLite database file path | `test.db` |
| `JWT_SECRET` | Secret key for JWT token signing (≥16 bytes) | `tian-niu-dev-secret-change-in-production` |
| `GIN_MODE` | Gin run mode (`debug`/`release`) | `debug` |

## Configuration

### Bash Tool Security

The built-in Bash tool executes shell commands with multiple security layers:

| Setting | Description | Default |
|---------|-------------|---------|
| `timeout_seconds` | Command execution timeout | `30` |
| `max_output_kb` | Maximum output size (truncated if exceeded) | `64` |
| `work_dir` | Restrict working directory (empty = no restriction) | `""` |
| `disabled` | Disable the Bash tool entirely | `false` |
| `allow_dangerous` | Bypass dangerous pattern blocking (use with caution) | `false` |

Security features:
- **Dangerous command blocking**: Patterns like `sudo`, `rm -rf /`, `mkfs`, `shutdown`, `curl | sh`, fork bombs, etc.
- **Sensitive environment filtering**: Environment variables with prefixes `JWT_`, `API_KEY`, `SECRET`, `TOKEN`, `PASSWORD`, `AWS_`, etc. are stripped from the command's environment
- **Output truncation**: Prevents memory exhaustion from large command output
- **Command length limit**: Maximum 4096 characters

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

- OpenAI: gpt-4o, gpt-4o-mini, gpt-4-turbo
- DeepSeek: deepseek-chat, deepseek-reasoner (with reasoning output)
- Zhipu AI: GLM-5.2, GLM-4
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
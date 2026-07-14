# TianNiu

> A lightweight AI chat agent backend built with Go, featuring streaming message processing, tool calls, MCP integration, and multi-level memory management.

[![Go Version](https://img.shields.io/badge/go-1.26.4-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)]()

## 🚀 Features

### Core Capabilities
- ✅ **Smart Conversations**: Fluent AI model interaction with streaming responses
- ✅ **Multi-thread Management**: Create, rename, and delete conversation threads
- ✅ **Streaming Messages**: Real-time message delivery via Server-Sent Events (SSE)
- ✅ **Tool Calls**: AI can invoke external tools to fetch information
- ✅ **MCP Integration**: Connect to any MCP-compatible tool server (stdio / HTTP)
- ✅ **Memory System**: Multi-level memory management (working/short/long-term)

### Security & Authentication
- ✅ **JWT Authentication**: User registration, login, and token-based access control
- ✅ **Bash Tool**: Execute shell commands with built-in security restrictions
- ✅ **Context Management**: Automatic message summarization and content offloading

### Extensibility
- ✅ **Skill Management**: Install, uninstall, and manage skills
- ✅ **MCP Server Management**: Install and manage MCP servers
- ✅ **Multi-database Support**: SQLite, PostgreSQL, and MySQL

### Planned Features
- [ ] **File Processing**: Support for file upload and analysis (RAG)
- [ ] **Message Editing**: Edit and recall sent messages
- [ ] **Message Search**: Search through message history
- [ ] **Conversation Export**: Export chat history in JSON/Markdown format
- [ ] **Parameter Tuning**: Customize temperature, max tokens, etc.
- [ ] **Web Search**: Real-time internet search capabilities
- [ ] **Multimodal Support**: Image generation and understanding
- [ ] **Voice Capabilities**: Speech-to-text and text-to-speech

## 🛠️ Tech Stack

### Backend
| Component | Technology |
|-----------|------------|
| Language | Go 1.26.4 |
| Framework | Gin |
| LLM SDK | OpenAI Go SDK v3 |
| MCP SDK | MCP Go SDK v1 |
| ORM | GORM |
| Database | SQLite / PostgreSQL / MySQL |
| Authentication | JWT (golang-jwt/v5) |

### Memory System
- **Working Memory**: In-memory circular buffer (max 100 messages)
- **Short-term Memory**: Database-backed with batch updates
- **Long-term Memory**: Vector database (PostgreSQL + pgvector) with semantic retrieval

## 📦 Installation

### Prerequisites
- Go 1.26.4+
- PostgreSQL 16+ (for production, with pgvector extension)
- Redis 7+ (for caching)

### Quick Start

1. **Clone the repository**
```bash
git clone https://github.com/tianniu-ai/tianniu.git
cd tianniu
```

2. **Install dependencies**
```bash
go mod download
```

3. **Configure the application**
```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml` with your LLM provider and database settings.

4. **Start the backend**
```bash
go run ./tianniu/main.go
```

The server runs on `http://localhost:8080`.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_TYPE` | Database type (`sqlite`, `postgres`, `mysql`) | `sqlite` |
| `DB_DSN` | Database connection string | `test.db` |
| `JWT_SECRET` | Secret key for JWT token signing (≥16 bytes) | `tian-niu-dev-secret-change-in-production` |
| `GIN_MODE` | Gin run mode (`debug`/`release`) | `debug` |
| `FRONT_MODEL_API_KEY` | API key for front LLM model | - |
| `BACK_MODEL_API_KEY` | API key for back LLM model (optional) | - |

## 🗄️ Database Configuration

### SQLite (Development)
```yaml
database:
  type: "sqlite"
  dsn: "test.db"
```

### PostgreSQL (Production)
```yaml
database:
  type: "postgres"
  dsn: "host=localhost port=5432 user=postgres password=postgres dbname=tianniu sslmode=disable"
```

### MySQL
```yaml
database:
  type: "mysql"
  dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
```

### Vector Database (Long-term Memory)

For long-term memory with semantic retrieval, enable the vector database:

```yaml
long_term_memory:
  enabled: true
  vector_db:
    host: localhost
    port: 5432
    user: admin
    password: password
    database: memory_db
    dimension: 1536
```

## 🔌 MCP Server Configuration

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

## 🧠 Supported Models

| Provider | Models |
|----------|--------|
| OpenAI | gpt-4o, gpt-4o-mini, gpt-4-turbo, gpt-5.2 |
| DeepSeek | deepseek-chat, deepseek-reasoner |
| Zhipu AI | GLM-5.2, GLM-4, GLM-4.6V |
| Qwen | QwQ, Qwen3 |
| Custom | Any model compatible with OpenAI API format |

## 🔌 API Documentation

For detailed API documentation, see [API Reference](doc/api_reference.md).

### Key Endpoints Summary

| Category | Endpoints |
|----------|-----------|
| **Authentication** | `/api/user/register`, `/api/user/login` |
| **Conversations** | `/api/conversation`, `/api/conversation/:id/message` |
| **Skills** | `/api/skills`, `/api/skills/:id/enable` |
| **MCP Servers** | `/api/mcps`, `/api/mcps/:id/enable` |
| **Health** | `/health`, `/metrics` |

## 📁 Project Structure

```
tianniu/
├── tianniu/                    # Main application entry
│   └── main.go                 # Application entry point
├── pkg/                        # Package directory
│   ├── agent/                  # Core agent logic
│   │   ├── agent.go            # Agent implementation
│   │   ├── manager.go          # Agent lifecycle management
│   │   ├── context/            # Conversation context
│   │   ├── memory/             # Memory system
│   │   ├── llm/                # LLM client
│   │   ├── skill/              # Skill system
│   │   ├── tool/               # Tool implementations
│   │   └── mcp/                # MCP integration
│   ├── repository/             # Data access layer
│   ├── server/                 # API server
│   ├── shared/                 # Shared utilities
│   └── rag/                    # RAG components
├── config.example.yaml         # Configuration template
├── doc/                        # Documentation
│   ├── overview.md             # System overview
│   ├── architecture.md         # Architecture design
│   ├── agent_module.md         # Agent module design
│   ├── memory_system.md        # Memory system design
│   └── ...                     # Other design documents
└── go.mod                      # Go dependencies
```

## 📚 Documentation

For comprehensive documentation, see the [Documentation Hub](doc/README.md).

| Document | Description |
|----------|-------------|
| `doc/README.md` | Documentation entry point |
| `doc/overview.md` | System overview and design principles |
| `doc/architecture.md` | High-level architecture design |
| `doc/agent_module.md` | Agent and context engine design |
| `doc/memory_system.md` | Multi-level memory system design |
| `doc/skill_system.md` | Skill system design |
| `doc/tool_system.md` | Tool system design |
| `doc/mcp_system.md` | MCP integration design |
| `doc/database_design.md` | Database schema design |
| `doc/api_design.md` | REST API design |
| `doc/deployment.md` | Deployment architecture |
| `doc/security.md` | Security considerations |
| `doc/monitoring.md` | Monitoring and observability |

## 🔧 Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o tianniu ./tianniu/main.go
```

### Docker
```bash
docker build -t tianniu .
docker run -p 8080:8080 tianniu
```

## 🤝 Contributing

We welcome contributions from the community!

### How to Contribute
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes (`git commit -am 'Add my feature'`)
4. Push to the branch (`git push origin feature/my-feature`)
5. Create a Pull Request

### Code Guidelines
- Follow Go coding standards (`go fmt`)
- Write unit tests for new functionality
- Add documentation for new features
- Use meaningful commit messages

## 📋 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Contact

For questions or support, please open an issue on GitHub or contact the development team.

---

**TianNiu** - Built with ❤️ for AI-powered conversations
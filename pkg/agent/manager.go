package agent

import (
	"sync"

	"github.com/openai/openai-go/v3"
	"github.com/tianniu-ai/tianniu/pkg/agent/context"
	"github.com/tianniu-ai/tianniu/pkg/agent/llm"
	"github.com/tianniu-ai/tianniu/pkg/agent/mcp"
	"github.com/tianniu-ai/tianniu/pkg/agent/memory"
	"github.com/tianniu-ai/tianniu/pkg/agent/tool"
	"github.com/tianniu-ai/tianniu/pkg/repository"
	"github.com/tianniu-ai/tianniu/pkg/shared"
)

type Manager struct {
	repo         *repository.Repository
	modelConf    shared.ModelConfig
	client       openai.Client
	tools        []tool.Tool
	systemPrompt string
	mcpClients   []*mcp.Client
	policies     []context.Policy
	memory       memory.Memory

	agents map[string]*Agent
	sync.RWMutex
}

func NewManager(
	repo *repository.Repository,
	modelConf shared.ModelConfig,
	systemPrompt string,
	tools []tool.Tool,
	mcpClients []*mcp.Client,
	policies []context.Policy,
	memory memory.Memory) *Manager {
	manger := &Manager{
		repo:         repo,
		modelConf:    modelConf,
		client:       llm.NewLLMClient(modelConf),
		tools:        tools,
		systemPrompt: systemPrompt,
		mcpClients:   mcpClients,
		policies:     policies,
		memory:       memory,
		agents:       make(map[string]*Agent),
	}
	return manger
}

func (m *Manager) GetAgent(userId, conversationId string) *Agent {
	m.RLock()
	agent, ok := m.agents[conversationId]
	if ok {
		m.RUnlock()
		return agent
	}
	m.RUnlock()

	m.Lock()
	defer m.Unlock()

	engine := context.NewContextEngine(m.memory, userId, conversationId, m.policies, m.repo)
	agent = NewAgent(m.modelConf, m.systemPrompt, m.tools, m.mcpClients, engine)
	m.agents[conversationId] = agent

	return agent
}

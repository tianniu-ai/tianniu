package agent

import (
	"sync"

	"github.com/openai/openai-go/v3"
	log "github.com/sirupsen/logrus"
	"github.com/tianniu-ai/tianniu/pkg/agent/context"
	"github.com/tianniu-ai/tianniu/pkg/agent/llm"
	"github.com/tianniu-ai/tianniu/pkg/agent/mcp"
	"github.com/tianniu-ai/tianniu/pkg/agent/memory"
	skill2 "github.com/tianniu-ai/tianniu/pkg/agent/skill"
	"github.com/tianniu-ai/tianniu/pkg/agent/tool"
	"github.com/tianniu-ai/tianniu/pkg/repository"
	"github.com/tianniu-ai/tianniu/pkg/shared"
)

type Manager struct {
	repo         *repository.SQLStore
	modelConf    shared.ModelConfig
	client       openai.Client
	tools        []tool.Tool
	systemPrompt string
	mcpClients   []*mcp.Client
	policies     []context.Policy
	memory       memory.Memory
	skillManager *skill2.Manager

	agents map[string]*Agent
	sync.RWMutex
}

func NewManager(
	repo *repository.SQLStore,
	modelConf shared.ModelConfig,
	systemPrompt string,
	tools []tool.Tool,
	mcpClients []*mcp.Client,
	policies []context.Policy,
	memory memory.Memory,
	skillManager *skill2.Manager) *Manager {
	manger := &Manager{
		repo:         repo,
		modelConf:    modelConf,
		client:       llm.NewLLMClient(modelConf),
		tools:        tools,
		systemPrompt: systemPrompt,
		mcpClients:   mcpClients,
		policies:     policies,
		memory:       memory,
		skillManager: skillManager,
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

	skillTools := m.loadSkillTools(userId)

	agent = NewAgent(m.modelConf, m.systemPrompt, m.tools, skillTools, m.mcpClients, engine)
	m.agents[conversationId] = agent

	return agent
}

func (m *Manager) loadSkillTools(userId string) []tool.Tool {
	if m.skillManager == nil {
		log.Warn("skill manager is nil, skipping skill tool loading")
		return nil
	}

	skills, err := m.skillManager.GetSkillsForUser(userId)
	if err != nil {
		log.Warnf("failed to get skills for user '%s': %v", userId, err)
		return nil
	}

	if len(skills) == 0 {
		log.Debugf("no skills found for user '%s'", userId)
		return nil
	}

	var tools []tool.Tool
	for _, s := range skills {
		if s.Status != skill2.SkillStatusEnabled {
			log.Debugf("skill '%s' is not enabled, skipping", s.Name)
			continue
		}
		skillTool := skill2.NewSkillTool(s, skill2.SkillToolConfig{})
		tools = append(tools, skillTool)
		log.Debugf("added skill tool '%s' for user '%s'", skillTool.ToolName(), userId)
	}

	log.Infof("loaded %d skill tools for user '%s'", len(tools), userId)
	return tools
}

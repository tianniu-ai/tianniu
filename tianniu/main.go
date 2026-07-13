package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jinzhu/configor"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/tianniu-ai/tianniu/pkg/agent"
	context2 "github.com/tianniu-ai/tianniu/pkg/agent/context"
	"github.com/tianniu-ai/tianniu/pkg/agent/mcp"
	"github.com/tianniu-ai/tianniu/pkg/agent/memory"
	skill2 "github.com/tianniu-ai/tianniu/pkg/agent/skill"
	"github.com/tianniu-ai/tianniu/pkg/agent/tool"
	"github.com/tianniu-ai/tianniu/pkg/repository"
	"github.com/tianniu-ai/tianniu/pkg/server"
	"github.com/tianniu-ai/tianniu/pkg/shared"
	_ "github.com/tianniu-ai/tianniu/pkg/shared/log"
)

type AppConfig struct {
	ServerAddress string `yaml:"server_address"`
	LLMProviders  struct {
		FrontModel shared.ModelConfig `yaml:"front_model"`
		BackModel  shared.ModelConfig `yaml:"back_model"`
	} `yaml:"llm_providers"`
	BashTool tool.BashToolConfig `yaml:"bash_tool"`
}

func loadAppConfig(path string) (AppConfig, error) {
	var config AppConfig
	err := configor.Load(&config, path)
	if err != nil {
		return AppConfig{}, err
	}
	return config, nil
}

func main() {
	_ = godotenv.Load()

	appConf, err := loadAppConfig("config.yaml")
	if err != nil {
		log.Errorf("Failed to load config.yaml: %v", err)
		panic(err)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "test.db"
	}
	db, err := repository.NewSQLStore(dbPath)
	if err != nil {
		log.Errorf("Failed to initialize database: %v", err)
		panic(err)
	}

	mcpServerMap, err := mcp.LoadMcpServerConfig("mcp-server.json")
	if err != nil {
		log.Errorf("Failed to load MCP server configuration: %v", err)
	}
	mcpClients := make([]*mcp.Client, 0)
	for k, v := range mcpServerMap {
		mcpClient := mcp.NewMcpToolProvider(k, v)
		if err := mcpClient.RefreshTools(context.Background()); err != nil {
			log.Errorf("Failed to refresh tools for MCP server %s: %v", k, err)
			continue
		}
		mcpClients = append(mcpClients, mcpClient)
	}

	// Create context engine and policies
	summarizer := context2.NewLLMSummarizer(appConf.LLMProviders.BackModel, 200)
	policies := []context2.Policy{
		context2.NewOffloadPolicy(db, 0.4, 0, 100),
		context2.NewSummaryPolicy(summarizer, 10, 20, 0.6),
		context2.NewTruncatePolicy(0, 0.85),
	}

	memoryUpdater := memory.NewLLMMemoryUpdater(appConf.LLMProviders.BackModel)
	multiLevelMemory := memory.NewMultiLevelMemory(db, memoryUpdater)

	skillsDir := os.Getenv("SKILLS_DIR")
	if skillsDir == "" {
		skillsDir = "skills"
	}

	skillStore := skill2.NewSQLSkillStore(db)

	skillManager := skill2.NewManager(skillStore, skillsDir)
	if err := skillManager.LoadInstalledSkills(); err != nil {
		log.Errorf("Failed to load installed skills: %v", err)
	}

	mgr := agent.NewManager(
		db,
		appConf.LLMProviders.FrontModel,
		agent.SystemPrompt,
		[]tool.Tool{tool.NewBashTool(appConf.BashTool)},
		mcpClients,
		policies,
		multiLevelMemory,
		skillManager)

	skillAPI := server.NewSkillAPI(skillManager)

	s := server.NewServer(appConf.ServerAddress, db, mgr, skillAPI)
	s.Run()
	defer s.Stop()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
}

package shared

import (
	"encoding/json"
	"os"
)

type AppConfig struct {
	LLMProviders struct {
		FrontModel ModelConfig `json:"front_model"`
	} `json:"llm_providers"`
	BashTool BashToolConfig `json:"bash_tool"`
}

type BashToolConfig struct {
	TimeoutSeconds int    `json:"timeout_seconds"`
	MaxOutputKB    int    `json:"max_output_kb"`
	WorkDir        string `json:"work_dir"`
	Disabled       bool   `json:"disabled"`
	AllowDangerous bool   `json:"allow_dangerous"`
}

type ModelConfig struct {
	BaseURL string `json:"base_url"`
	ApiKey  string `json:"api_key"`
	Model   string `json:"model"`

	ContextWindow int `json:"context_window"`
}

func LoadAppConfig(path string) (AppConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, err
	}
	var config AppConfig
	err = json.Unmarshal(content, &config)
	if err != nil {
		return AppConfig{}, err
	}
	return config, nil
}

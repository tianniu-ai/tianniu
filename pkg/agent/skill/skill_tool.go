package skill

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
	"github.com/tianniu-ai/tianniu/pkg/agent/tool"
)

type SkillTool struct {
	skill  *Skill
	config SkillToolConfig
}

type SkillToolConfig struct {
	TimeoutSeconds int    `yaml:"timeout_seconds"`
	WorkDir        string `yaml:"work_dir"`
}

func NewSkillTool(skill *Skill, config SkillToolConfig) *SkillTool {
	if config.TimeoutSeconds <= 0 {
		config.TimeoutSeconds = 30
	}
	return &SkillTool{
		skill:  skill,
		config: config,
	}
}

func (t *SkillTool) ToolName() tool.AgentTool {
	return tool.AgentTool(fmt.Sprintf("skill_%s", t.skill.Name))
}

func (t *SkillTool) Info() openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
		Name:        t.ToolName(),
		Description: openai.String(t.skill.Description),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "the command to execute (use 'info' to get skill information)",
				},
				"args": map[string]any{
					"type":        "object",
					"description": "arguments for the command (key-value pairs)",
				},
			},
			"required": []string{"command"},
		},
	})
}

type SkillToolParams struct {
	Command string                 `json:"command"`
	Args    map[string]interface{} `json:"args,omitempty"`
}

func (t *SkillTool) Execute(ctx context.Context, argumentsInJSON string) (string, error) {
	if t.skill.Status != SkillStatusEnabled {
		return "", fmt.Errorf("skill '%s' is not enabled", t.skill.Name)
	}

	var params SkillToolParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if params.Command == "" {
		return "", fmt.Errorf("command is required")
	}

	return t.skill.Content, nil
}

func (t *SkillTool) GetSkill() *Skill {
	return t.skill
}

func (t *SkillTool) GetConfig() SkillToolConfig {
	return t.config
}

func (t *SkillTool) UpdateConfig(config SkillToolConfig) {
	t.config = config
}

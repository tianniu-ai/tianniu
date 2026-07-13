package skill

import (
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
)

type SkillStatus string

const (
	SkillStatusInstalled SkillStatus = "installed"
	SkillStatusEnabled   SkillStatus = "enabled"
	SkillStatusDisabled  SkillStatus = "disabled"
	SkillStatusRemoved   SkillStatus = "removed"
)

type SkillType string

const (
	SkillTypeSystem SkillType = "system"
	SkillTypeUser   SkillType = "user"
)

type SkillMetadata struct {
	Emoji    string                 `json:"emoji"`
	Requires map[string][]string    `json:"requires,omitempty"`
	Author   string                 `json:"author,omitempty"`
	Version  string                 `json:"version,omitempty"`
	License  string                 `json:"license,omitempty"`
	Category string                 `json:"category,omitempty"`
	Tags     []string               `json:"tags,omitempty"`
	Homepage string                 `json:"homepage,omitempty"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
}

type SkillDefinition struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Homepage     string                 `json:"homepage"`
	Metadata     SkillMetadata          `json:"metadata"`
	Commands     []SkillCommand         `json:"commands"`
	WhenToUse    []string               `json:"when_to_use"`
	WhenNotToUse []string               `json:"when_not_to_use"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	FunctionDef  map[string]interface{} `json:"function_def,omitempty"`
}

type SkillCommand struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Command      string                 `json:"command"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	OutputFormat string                 `json:"output_format,omitempty"`
}

type Skill struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Homepage    string           `json:"homepage"`
	Metadata    SkillMetadata    `json:"metadata"`
	Status      SkillStatus      `json:"status"`
	Type        SkillType        `json:"type"`
	UserID      string           `json:"user_id"`
	InstalledAt time.Time        `json:"installed_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Path        string           `json:"path"`
	Definition  *SkillDefinition `json:"definition"`
	Content     string           `json:"content"`
}

func (s *Skill) ToMarkdown() string {
	var md strings.Builder

	if s.Metadata.Emoji != "" {
		md.WriteString(s.Metadata.Emoji + " ")
	}
	md.WriteString("# " + s.Name + "\n\n")

	if s.Description != "" {
		md.WriteString("## Description\n\n" + s.Description + "\n\n")
	}

	if s.Homepage != "" {
		md.WriteString("**Homepage:** [" + s.Homepage + "](" + s.Homepage + ")\n\n")
	}

	if s.Metadata.Author != "" {
		md.WriteString("**Author:** " + s.Metadata.Author + "\n")
	}
	if s.Metadata.Version != "" {
		md.WriteString("**Version:** " + s.Metadata.Version + "\n")
	}
	if s.Metadata.Category != "" {
		md.WriteString("**Category:** " + s.Metadata.Category + "\n")
	}
	if len(s.Metadata.Tags) > 0 {
		md.WriteString("**Tags:** " + strings.Join(s.Metadata.Tags, ", ") + "\n")
	}
	if s.Metadata.Author != "" || s.Metadata.Version != "" || s.Metadata.Category != "" || len(s.Metadata.Tags) > 0 {
		md.WriteString("\n")
	}

	if s.Definition != nil {
		if len(s.Definition.WhenToUse) > 0 {
			md.WriteString("## When to Use\n\n")
			for _, item := range s.Definition.WhenToUse {
				md.WriteString("- " + item + "\n")
			}
			md.WriteString("\n")
		}

		if len(s.Definition.WhenNotToUse) > 0 {
			md.WriteString("## When NOT to Use\n\n")
			for _, item := range s.Definition.WhenNotToUse {
				md.WriteString("- " + item + "\n")
			}
			md.WriteString("\n")
		}

		if len(s.Definition.Commands) > 0 {
			md.WriteString("## Commands\n\n")
			for _, cmd := range s.Definition.Commands {
				md.WriteString("### " + cmd.Name + "\n\n")
				if cmd.Description != "" {
					md.WriteString(cmd.Description + "\n\n")
				}
				if cmd.Command != "" {
					md.WriteString("```\n" + cmd.Command + "\n```\n\n")
				}
			}
		}
	}

	md.WriteString("**Status:** " + string(s.Status) + "\n")
	md.WriteString("**Type:** " + string(s.Type) + "\n")

	return md.String()
}

type SkillToolInterface interface {
	ToolName() string
	Info() openai.ChatCompletionToolUnionParam
	Execute(ctx interface{}, argumentsInJSON string) (string, error)
}

type InstallOptions struct {
	Force bool `json:"force"`
}

type UninstallOptions struct {
	KeepConfig bool `json:"keep_config"`
}

type SkillStore interface {
	GetAll() ([]*Skill, error)
	GetByID(id string) (*Skill, error)
	GetByName(name string) (*Skill, error)
	GetByUserID(userID string) ([]*Skill, error)
	GetSystemSkills() ([]*Skill, error)
	GetUserSkills(userID string) ([]*Skill, error)
	GetSkillForUser(userID, skillName string) (*Skill, error)
	Save(skill *Skill) error
	Delete(id string) error
	UpdateStatus(id string, status SkillStatus) error
}

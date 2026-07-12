package skill

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tianniu-ai/tianniu/pkg/agent/tool"
)

type Manager struct {
	store       SkillStore
	skillsDir   string
	nativeTools map[tool.AgentTool]tool.Tool
}

func NewManager(store SkillStore, skillsDir string) *Manager {
	return &Manager{
		store:       store,
		skillsDir:   skillsDir,
		nativeTools: make(map[tool.AgentTool]tool.Tool),
	}
}

func (m *Manager) InstallSystemSkill(skillPath string, options InstallOptions) (*Skill, error) {
	if !filepath.IsAbs(skillPath) {
		absPath, err := filepath.Abs(skillPath)
		if err != nil {
			return nil, fmt.Errorf("invalid skill path: %w", err)
		}
		skillPath = absPath
	}

	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("skill path does not exist: %s", skillPath)
	}

	skillDef, content, err := parseSkillDefinition(skillPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse skill definition: %w", err)
	}

	log.Infof("Install system skill: %s", skillDef.Name)

	existingSkill, _ := m.store.GetByName(skillDef.Name)
	if existingSkill != nil && !options.Force {
		return nil, fmt.Errorf("system skill '%s' is already installed", skillDef.Name)
	}

	destPath := filepath.Join(m.skillsDir, "system", skillDef.Name)
	if err := copyDir(skillPath, destPath); err != nil {
		return nil, fmt.Errorf("failed to copy skill files: %w", err)
	}

	skill := &Skill{
		ID:          uuid.NewString(),
		Name:        skillDef.Name,
		Description: skillDef.Description,
		Homepage:    skillDef.Homepage,
		Metadata:    skillDef.Metadata,
		Status:      SkillStatusEnabled,
		Type:        SkillTypeSystem,
		UserID:      "",
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
		Path:        destPath,
		Definition:  skillDef,
		Content:     content,
	}

	if existingSkill != nil {
		skill.ID = existingSkill.ID
	}

	if err := m.store.Save(skill); err != nil {
		return nil, fmt.Errorf("failed to save skill: %w", err)
	}

	log.Infof("System skill '%s' installed successfully", skill.Name)
	return skill, nil
}

func (m *Manager) InstallUserSkill(userID, skillPath string, options InstallOptions) (*Skill, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	if !filepath.IsAbs(skillPath) {
		absPath, err := filepath.Abs(skillPath)
		if err != nil {
			return nil, fmt.Errorf("invalid skill path: %w", err)
		}
		skillPath = absPath
	}

	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("skill path does not exist: %s", skillPath)
	}

	skillDef, content, err := parseSkillDefinition(skillPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse skill definition: %w", err)
	}

	existingSkill, _ := m.store.GetSkillForUser(userID, skillDef.Name)
	if existingSkill != nil && existingSkill.Type == SkillTypeUser && !options.Force {
		return nil, fmt.Errorf("user skill '%s' is already installed", skillDef.Name)
	}

	destPath := filepath.Join(m.skillsDir, "users", userID, skillDef.Name)
	if err := copyDir(skillPath, destPath); err != nil {
		return nil, fmt.Errorf("failed to copy skill files: %w", err)
	}

	skill := &Skill{
		ID:          uuid.NewString(),
		Name:        skillDef.Name,
		Description: skillDef.Description,
		Homepage:    skillDef.Homepage,
		Metadata:    skillDef.Metadata,
		Status:      SkillStatusEnabled,
		Type:        SkillTypeUser,
		UserID:      userID,
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
		Path:        destPath,
		Definition:  skillDef,
		Content:     content,
	}

	if existingSkill != nil && existingSkill.Type == SkillTypeUser {
		skill.ID = existingSkill.ID
	}

	if err := m.store.Save(skill); err != nil {
		return nil, fmt.Errorf("failed to save skill: %w", err)
	}

	log.Infof("User skill '%s' installed successfully for user '%s'", skill.Name, userID)
	return skill, nil
}

func (m *Manager) InstallUserSkillFromContent(userID, content string, options InstallOptions) (*Skill, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	skillDef, skillContent, err := parseSkillMD(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse skill definition: %w", err)
	}

	existingSkill, _ := m.store.GetSkillForUser(userID, skillDef.Name)
	if existingSkill != nil && existingSkill.Type == SkillTypeUser && !options.Force {
		return nil, fmt.Errorf("user skill '%s' is already installed", skillDef.Name)
	}

	destPath := filepath.Join(m.skillsDir, "users", userID, skillDef.Name)
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create skill directory: %w", err)
	}

	skillMDPath := filepath.Join(destPath, "SKILL.md")
	if err := os.WriteFile(skillMDPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write SKILL.md: %w", err)
	}

	skill := &Skill{
		ID:          uuid.NewString(),
		Name:        skillDef.Name,
		Description: skillDef.Description,
		Homepage:    skillDef.Homepage,
		Metadata:    skillDef.Metadata,
		Status:      SkillStatusEnabled,
		Type:        SkillTypeUser,
		UserID:      userID,
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
		Path:        destPath,
		Definition:  skillDef,
		Content:     skillContent,
	}

	if existingSkill != nil && existingSkill.Type == SkillTypeUser {
		skill.ID = existingSkill.ID
	}

	if err := m.store.Save(skill); err != nil {
		return nil, fmt.Errorf("failed to save skill: %w", err)
	}

	log.Infof("User skill '%s' installed successfully for user '%s'", skill.Name, userID)
	return skill, nil
}

func (m *Manager) Uninstall(skillID string, options UninstallOptions) error {
	skill, err := m.store.GetByID(skillID)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(skill.Path); err != nil {
		return fmt.Errorf("failed to remove skill directory: %w", err)
	}

	if err := m.store.Delete(skillID); err != nil {
		return fmt.Errorf("failed to delete skill from store: %w", err)
	}

	log.Infof("Skill '%s' uninstalled successfully", skill.Name)
	return nil
}

func (m *Manager) Enable(skillID string) error {
	return m.store.UpdateStatus(skillID, SkillStatusEnabled)
}

func (m *Manager) Disable(skillID string) error {
	return m.store.UpdateStatus(skillID, SkillStatusDisabled)
}

func (m *Manager) GetAllSkills() ([]*Skill, error) {
	return m.store.GetAll()
}

func (m *Manager) GetSkillByID(id string) (*Skill, error) {
	return m.store.GetByID(id)
}

func (m *Manager) GetSkillByName(name string) (*Skill, error) {
	return m.store.GetByName(name)
}

func (m *Manager) GetSystemSkills() ([]*Skill, error) {
	return m.store.GetSystemSkills()
}

func (m *Manager) GetUserSkills(userID string) ([]*Skill, error) {
	return m.store.GetUserSkills(userID)
}

func (m *Manager) GetSkillsForUser(userID string) ([]*Skill, error) {
	systemSkills, err := m.store.GetSystemSkills()
	if err != nil {
		return nil, err
	}

	userSkills, err := m.store.GetUserSkills(userID)
	if err != nil {
		return nil, err
	}

	allSkills := append(systemSkills, userSkills...)
	return allSkills, nil
}

func (m *Manager) GetSkillForUser(userID, skillName string) (*Skill, error) {
	return m.store.GetSkillForUser(userID, skillName)
}

func (m *Manager) LoadInstalledSkills() error {
	if err := os.MkdirAll(filepath.Join(m.skillsDir, "system"), 0755); err != nil {
		return fmt.Errorf("failed to create system skills directory: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(m.skillsDir, "users"), 0755); err != nil {
		return fmt.Errorf("failed to create users skills directory: %w", err)
	}

	if err := m.loadSystemSkills(); err != nil {
		log.Warnf("Failed to load system skills: %v", err)
	}

	return nil
}

func (m *Manager) loadSystemSkills() error {
	systemDir := filepath.Join(m.skillsDir, "system")
	entries, err := os.ReadDir(systemDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read system skills directory: %w", err)
	}

	currentSkillNames := make(map[string]bool)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillPath := filepath.Join(systemDir, entry.Name())
		skillDef, content, err := parseSkillDefinition(skillPath)
		if err != nil {
			log.Warnf("Failed to parse system skill '%s': %v", entry.Name(), err)
			continue
		}

		currentSkillNames[skillDef.Name] = true

		existingSkill, err := m.store.GetByName(skillDef.Name)
		if err != nil || existingSkill == nil {
			skill := &Skill{
				ID:          uuid.NewString(),
				Name:        skillDef.Name,
				Description: skillDef.Description,
				Homepage:    skillDef.Homepage,
				Metadata:    skillDef.Metadata,
				Status:      SkillStatusEnabled,
				Type:        SkillTypeSystem,
				UserID:      "",
				InstalledAt: time.Now(),
				UpdatedAt:   time.Now(),
				Path:        skillPath,
				Definition:  skillDef,
				Content:     content,
			}

			if err := m.store.Save(skill); err != nil {
				log.Warnf("Failed to save system skill '%s': %v", skill.Name, err)
			} else {
				log.Infof("Installed system skill '%s'", skill.Name)
			}
			continue
		}

		if existingSkill.Type != SkillTypeSystem {
			continue
		}

		needUpdate := false
		if existingSkill.Description != skillDef.Description ||
			existingSkill.Homepage != skillDef.Homepage ||
			existingSkill.Path != skillPath {
			needUpdate = true
		}

		if needUpdate {
			existingSkill.Description = skillDef.Description
			existingSkill.Homepage = skillDef.Homepage
			existingSkill.Metadata = skillDef.Metadata
			existingSkill.Path = skillPath
			existingSkill.Definition = skillDef
			existingSkill.Content = content
			existingSkill.UpdatedAt = time.Now()

			if err := m.store.Save(existingSkill); err != nil {
				log.Warnf("Failed to update system skill '%s': %v", existingSkill.Name, err)
			} else {
				log.Infof("Updated system skill '%s'", existingSkill.Name)
			}
		}
	}

	existingSkills, err := m.store.GetSystemSkills()
	if err != nil {
		log.Warnf("Failed to get existing system skills: %v", err)
	} else {
		for _, existingSkill := range existingSkills {
			if !currentSkillNames[existingSkill.Name] {
				if err := m.store.Delete(existingSkill.ID); err != nil {
					log.Warnf("Failed to remove missing system skill '%s': %v", existingSkill.Name, err)
				} else {
					log.Infof("Removed missing system skill '%s'", existingSkill.Name)
				}
			}
		}
	}

	return nil
}

func parseSkillDefinition(skillPath string) (*SkillDefinition, string, error) {
	skillMDPath := filepath.Join(skillPath, "SKILL.md")
	data, err := os.ReadFile(skillMDPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read SKILL.md: %w", err)
	}

	def, content, err := parseSkillMD(string(data))
	if err != nil {
		return nil, "", err
	}

	return def, content, nil
}

func parseSkillMD(content string) (*SkillDefinition, string, error) {
	lines := strings.Split(content, "\n")
	def := &SkillDefinition{
		Commands:     []SkillCommand{},
		WhenToUse:    []string{},
		WhenNotToUse: []string{},
	}

	var markdownBuffer strings.Builder

	inMetadata := false
	metadataEnded := false
	inWhenToUse := false
	inWhenNotToUse := false
	var metadataBuffer bytes.Buffer

	for _, line := range lines {
		if strings.HasPrefix(line, "---") {
			if inMetadata {
				inMetadata = false
				metadataEnded = true
				var metadata SkillMetadata
				if err := parseYAML(metadataBuffer.String(), &metadata); err != nil {
					return nil, "", fmt.Errorf("failed to parse metadata: %w", err)
				}
				def.Metadata = metadata
			} else {
				inMetadata = true
				metadataBuffer.Reset()
			}
			continue
		}

		if inMetadata {
			metadataBuffer.WriteString(line + "\n")

			if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "name":
					def.Name = value
				case "description":
					def.Description = value
				case "homepage":
					def.Homepage = value
				}
			}
			continue
		}

		if metadataEnded {
			markdownBuffer.WriteString(line + "\n")
		}

		if strings.HasPrefix(line, "## When to Use") {
			inWhenToUse = true
			inWhenNotToUse = false
			continue
		}

		if strings.HasPrefix(line, "## When NOT to Use") {
			inWhenNotToUse = true
			inWhenToUse = false
			continue
		}

		if inWhenToUse && strings.HasPrefix(line, "- ") {
			def.WhenToUse = append(def.WhenToUse, strings.TrimPrefix(line, "- "))
			continue
		}

		if inWhenNotToUse && strings.HasPrefix(line, "- ") {
			def.WhenNotToUse = append(def.WhenNotToUse, strings.TrimPrefix(line, "- "))
			continue
		}

		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "name":
				def.Name = value
			case "description":
				def.Description = value
			case "homepage":
				def.Homepage = value
			}
		}
	}

	if def.Name == "" {
		return nil, "", fmt.Errorf("skill definition is missing 'name' field")
	}

	return def, strings.TrimSpace(markdownBuffer.String()), nil
}

func parseYAML(content string, v interface{}) error {
	return parseSimpleYAML(content, v)
}

func parseSimpleYAML(content string, v interface{}) error {
	lines := strings.Split(content, "\n")
	result := make(map[string]interface{})

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
			var jsonValue map[string]interface{}
			if err := parseSimpleJSON(value, &jsonValue); err == nil {
				result[key] = jsonValue
			} else {
				result[key] = value
			}
		} else {
			result[key] = value
		}
	}

	data, _ := json.Marshal(result)
	return json.Unmarshal(data, v)
}

func parseSimpleJSON(content string, v interface{}) error {
	return nil
}

func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}

			if err := os.WriteFile(dstPath, data, entry.Type().Perm()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) GetNativeTools() map[tool.AgentTool]tool.Tool {
	return m.nativeTools
}

func (m *Manager) AddNativeTool(t tool.Tool) {
	m.nativeTools[t.ToolName()] = t
}

func (m *Manager) RemoveNativeTool(name tool.AgentTool) {
	delete(m.nativeTools, name)
}

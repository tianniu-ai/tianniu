package skill

import (
	"github.com/tianniu-ai/tianniu/pkg/model"
	"github.com/tianniu-ai/tianniu/pkg/repository"
)

// SQLSkillStore implements SkillStore interface using SQLStore
type SQLSkillStore struct {
	repo *repository.SQLStore
}

// NewSQLSkillStore creates a new SQLSkillStore
func NewSQLSkillStore(repo *repository.SQLStore) *SQLSkillStore {
	return &SQLSkillStore{repo: repo}
}

func (s *SQLSkillStore) GetAll() ([]*Skill, error) {
	modelSkills, err := s.repo.GetAllSkills()
	if err != nil {
		return nil, err
	}
	return convertModelSkillsToSkill(modelSkills), nil
}

func (s *SQLSkillStore) GetByID(id string) (*Skill, error) {
	modelSkill, err := s.repo.GetSkillByID(id)
	if err != nil {
		return nil, err
	}
	return convertModelSkillToSkill(modelSkill), nil
}

func (s *SQLSkillStore) GetByName(name string) (*Skill, error) {
	modelSkill, err := s.repo.GetSkillByName(name)
	if err != nil {
		return nil, err
	}
	return convertModelSkillToSkill(modelSkill), nil
}

func (s *SQLSkillStore) GetByUserID(userID string) ([]*Skill, error) {
	modelSkills, err := s.repo.GetSkillsByUserID(userID)
	if err != nil {
		return nil, err
	}
	return convertModelSkillsToSkill(modelSkills), nil
}

func (s *SQLSkillStore) GetSystemSkills() ([]*Skill, error) {
	modelSkills, err := s.repo.GetSystemSkills()
	if err != nil {
		return nil, err
	}
	return convertModelSkillsToSkill(modelSkills), nil
}

func (s *SQLSkillStore) GetUserSkills(userID string) ([]*Skill, error) {
	modelSkills, err := s.repo.GetUserSkills(userID)
	if err != nil {
		return nil, err
	}
	return convertModelSkillsToSkill(modelSkills), nil
}

func (s *SQLSkillStore) GetSkillForUser(userID, skillName string) (*Skill, error) {
	modelSkill, err := s.repo.GetSkillForUser(userID, skillName)
	if err != nil {
		return nil, err
	}
	return convertModelSkillToSkill(modelSkill), nil
}

func (s *SQLSkillStore) Save(skill *Skill) error {
	modelSkill := convertSkillToModelSkill(skill)
	return s.repo.SaveSkill(modelSkill)
}

func (s *SQLSkillStore) Delete(id string) error {
	return s.repo.DeleteSkill(id)
}

func (s *SQLSkillStore) UpdateStatus(id string, status SkillStatus) error {
	return s.repo.UpdateSkillStatus(id, string(status))
}

// convertSkillToModelSkill converts Skill to model.Skill
func convertSkillToModelSkill(s *Skill) *model.Skill {
	return &model.Skill{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Homepage:    s.Homepage,
		Metadata: model.SkillMetadata{
			Emoji:    s.Metadata.Emoji,
			Author:   s.Metadata.Author,
			Version:  s.Metadata.Version,
			License:  s.Metadata.License,
			Category: s.Metadata.Category,
			Homepage: s.Metadata.Homepage,
		},
		Status:      string(s.Status),
		Type:        string(s.Type),
		UserID:      s.UserID,
		InstalledAt: s.InstalledAt,
		UpdatedAt:   s.UpdatedAt,
		Path:        s.Path,
		Content:     s.Content,
	}
}

// convertModelSkillToSkill converts model.Skill to Skill
func convertModelSkillToSkill(s *model.Skill) *Skill {
	return &Skill{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Homepage:    s.Homepage,
		Metadata: SkillMetadata{
			Emoji:    s.Metadata.Emoji,
			Author:   s.Metadata.Author,
			Version:  s.Metadata.Version,
			License:  s.Metadata.License,
			Category: s.Metadata.Category,
			Homepage: s.Metadata.Homepage,
		},
		Status:      SkillStatus(s.Status),
		Type:        SkillType(s.Type),
		UserID:      s.UserID,
		InstalledAt: s.InstalledAt,
		UpdatedAt:   s.UpdatedAt,
		Path:        s.Path,
		Content:     s.Content,
	}
}

// convertModelSkillsToSkill converts a slice of model.Skill to Skill pointers
func convertModelSkillsToSkill(skills []*model.Skill) []*Skill {
	result := make([]*Skill, 0, len(skills))
	for _, s := range skills {
		result = append(result, convertModelSkillToSkill(s))
	}
	return result
}

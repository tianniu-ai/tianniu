package repository

import (
	"github.com/tianniu-ai/tianniu/pkg/model"
)

func (r *Repository) GetConversationByID(id string) (*model.Conversation, error) {
	var m model.Conversation
	if err := r.db.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetUserConversations(userID string) ([]*model.Conversation, error) {
	var convs []*model.Conversation
	if err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&convs).Error; err != nil {
		return nil, err
	}
	return convs, nil
}

func (r *Repository) UpdateConversationTitle(conversation *model.Conversation) error {
	if err := r.db.Model(&model.Conversation{}).
		Where("id = ?", conversation.ID).
		Update("title", conversation.Title).Error; err != nil {
		return err
	}
	return nil
}

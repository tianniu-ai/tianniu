package repository

import "github.com/tianniu-ai/tianniu/pkg/model"

func (r *Repository) GetConversationMessages(conversationID string, limit int) ([]*model.ChatMessage, error) {
	var list []*model.ChatMessage
	query := r.db.Where("conversation_id = ?", conversationID).Order("created_at desc")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

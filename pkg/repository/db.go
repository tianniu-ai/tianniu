package repository

import (
	"errors"
	"strings"

	"github.com/libtnb/sqlite"
	"github.com/tianniu-ai/tianniu/pkg/model"
	"gorm.io/gorm"
)

var ErrDuplicateEntry = errors.New("duplicate entry")

type Repository struct {
	db *gorm.DB
}

func NewRepository(dsn string) (*Repository, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&model.User{}, &model.Conversation{}, &model.ChatMessage{}, &model.KVData{})
	if err != nil {
		return nil, err
	}
	return &Repository{
		db: db,
	}, nil
}

func (r *Repository) Create(v interface{}) error {
	err := r.db.Create(v).Error
	if err != nil && isDuplicateEntryError(err) {
		return ErrDuplicateEntry
	}
	return err
}

func (r *Repository) Delete(v interface{}) error {
	return r.db.Delete(v).Error
}

func (r *Repository) DeleteConversationWithMessages(conversationID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("conversation_id = ?", conversationID).Delete(&model.ChatMessage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", conversationID).Delete(&model.Conversation{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func isDuplicateEntryError(err error) bool {
	s := err.Error()
	return strings.Contains(s, "UNIQUE constraint failed") || strings.Contains(s, "duplicate key")
}

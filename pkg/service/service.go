package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/tianniu-ai/tianniu/pkg/auth"
	"github.com/tianniu-ai/tianniu/pkg/model"
	"github.com/tianniu-ai/tianniu/pkg/repository"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tianniu-ai/tianniu/pkg/agent"
	"github.com/tianniu-ai/tianniu/pkg/shared"
	"github.com/tianniu-ai/tianniu/pkg/vo"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db  *repository.Repository
	mgr *agent.Manager
}

func NewService(db *repository.Repository, mgr *agent.Manager) *Service {
	return &Service{db: db, mgr: mgr}
}

func (s *Service) Register(req vo.RegisterReq) (vo.UserVO, error) {
	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return vo.UserVO{}, err
	}

	user := &model.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now().Unix(),
	}

	err = s.db.Create(user)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEntry) {
			return vo.UserVO{}, errors.New("username already exists")
		}
		return vo.UserVO{}, err
	}
	return vo.UserVO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *Service) Login(req vo.LoginReq) (vo.LoginRespVO, error) {

	user, err := s.db.GetUserByUsername(req.Username)
	if err != nil {
		return vo.LoginRespVO{}, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return vo.LoginRespVO{}, err
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		return vo.LoginRespVO{}, err
	}

	return vo.LoginRespVO{
		User: vo.UserVO{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		Token: token,
	}, nil
}

func (s *Service) CreateConversation(req vo.CreateConversationReq) (vo.ConversationVO, error) {
	conv := model.Conversation{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Title:     req.Title,
		CreatedAt: time.Now().Unix(),
	}
	if err := s.db.Create(&conv); err != nil {
		return vo.ConversationVO{}, err
	}
	return vo.ConversationVO{
		ID:        conv.ID,
		UserID:    conv.UserID,
		Title:     conv.Title,
		CreatedAt: conv.CreatedAt,
	}, nil
}

func (s *Service) ListConversations(userID string) ([]vo.ConversationVO, error) {

	list, err := s.db.GetUserConversations(userID)
	if err != nil {
		return nil, err
	}
	result := make([]vo.ConversationVO, 0, len(list))
	for _, conv := range list {
		result = append(result, vo.ConversationVO{
			ID:        conv.ID,
			UserID:    conv.UserID,
			Title:     conv.Title,
			CreatedAt: conv.CreatedAt,
		})
	}
	return result, nil
}

func (s *Service) RenameConversation(userId, conversationID string, title string) (vo.ConversationVO, error) {

	conv, err := s.db.GetConversationByID(conversationID)
	if err != nil {
		return vo.ConversationVO{}, err
	}
	if conv.UserID != userId {
		return vo.ConversationVO{}, errors.New("user_id must match the one in the token")
	}

	conv.Title = title
	if err := s.db.UpdateConversationTitle(conv); err != nil {
		return vo.ConversationVO{}, err
	}

	return vo.ConversationVO{
		ID:        conv.ID,
		UserID:    conv.UserID,
		Title:     conv.Title,
		CreatedAt: conv.CreatedAt,
	}, nil
}

func (s *Service) DeleteConversation(userId, conversationID string) error {
	conv, err := s.db.GetConversationByID(conversationID)
	if err != nil {
		return err
	}
	if conv.UserID != userId {
		return errors.New("user_id must match the one in the token")
	}

	if err := s.db.DeleteConversationWithMessages(conversationID); err != nil {
		return err
	}
	return nil
}

func (s *Service) ListMessages(userID, conversationID string) ([]vo.ChatMessageVO, error) {
	// Verify conversation belongs to the authenticated user
	conv, err := s.db.GetConversationByID(conversationID)
	if err != nil {
		return nil, err
	}
	if conv.UserID != userID {
		return nil, errors.New("user_id must match the one in the token")
	}

	msgs, err := s.db.GetConversationMessages(conversationID, -1)
	if err != nil {
		return nil, err
	}

	result := make([]vo.ChatMessageVO, 0, len(msgs))
	for _, msg := range msgs {
		result = append(result, vo.ChatMessageVO{
			MessageID:       msg.ID,
			ConversationID:  msg.ConversationID,
			ParentMessageID: msg.ParentMessageID,
			Query:           msg.Query,
			Response:        msg.Response,
			Model:           msg.Model,
			CreatedAt:       msg.CreatedAt,
			Rounds:          parseRounds(msg.Rounds),
		})
	}
	return result, nil
}

// CreateMessage validates conversation, builds history, saves message record, and starts agent streaming execution.
func (s *Service) CreateMessage(ctx context.Context, conversationID string, req vo.CreateMessageReq, voCh chan<- vo.SSEMessageVO) error {
	// Validate conversation exists
	conv, err := s.db.GetConversationByID(conversationID)
	if err != nil {
		return err
	}
	if conv.UserID != req.UserID {
		return errors.New("user_id must match the one in the token")
	}

	msgID := uuid.New().String()
	createdAt := time.Now().Unix()

	eventCh := make(chan agent.StreamEvent, 64)
	defer close(eventCh)

	// Bridge agent events -> SSE events with non-blocking send to avoid
	// goroutine leak when the client disconnects mid-stream.
	go func() {
		for e := range eventCh {
			select {
			case <-ctx.Done():
				return
			case voCh <- toSSEMessage(msgID, e):
			default:
				return
			}
		}
	}()

	agent := s.mgr.GetAgent(req.UserID, conversationID)

	// TODO: ParentMessageID is not used yet.
	result, runErr := agent.RunStreaming(ctx, req.Query, eventCh)
	if runErr != nil {
		log.Warnf("run streaming error: %v", runErr)
		return runErr
	}

	roundsJSON, _ := json.Marshal(result.Rounds)
	usageJSON, _ := json.Marshal(result.Usage)
	s.db.Create(&model.ChatMessage{
		ID:              msgID,
		UserID:          req.UserID,
		ConversationID:  conversationID,
		ParentMessageID: req.ParentMessageID,
		Query:           req.Query,
		Response:        result.Response,
		Rounds:          string(roundsJSON),
		Usage:           string(usageJSON),
		Model:           agent.Model(),
		CreatedAt:       createdAt,
	})

	return nil
}

func toSSEMessage(msgID string, e agent.StreamEvent) vo.SSEMessageVO {
	msg := vo.SSEMessageVO{MessageID: msgID, Event: e.Event}
	switch e.Event {
	case agent.EventReasoning:
		msg.ReasoningContent = &e.ReasoningContent
	case agent.EventContent, agent.EventError:
		msg.Content = &e.Content
	case agent.EventToolCall:
		msg.ToolCall = &e.ToolCall
		msg.ToolArguments = &e.ToolArguments
	case agent.EventToolResult:
		msg.ToolCall = &e.ToolCall
		msg.ToolResult = &e.ToolResult
	}
	return msg
}

// parseRounds converts stored rounds JSON to frontend-friendly RoundMessageVO list.
func parseRounds(roundsJSON string) []vo.RoundMessageVO {
	if roundsJSON == "" {
		return nil
	}
	var msgs []shared.OpenAIMessage
	if err := json.Unmarshal([]byte(roundsJSON), &msgs); err != nil {
		return nil
	}

	result := make([]vo.RoundMessageVO, 0, len(msgs))
	for _, m := range msgs {
		switch {
		case m.OfUser != nil:
			// user messages don't need to be displayed
			continue

		case m.OfAssistant != nil:
			a := m.OfAssistant
			rv := vo.RoundMessageVO{Role: "assistant"}
			if len(a.ToolCalls) > 0 {
				for _, tc := range a.ToolCalls {
					if tc.OfFunction != nil {
						rv.ToolCalls = append(rv.ToolCalls, vo.ToolCallVO{
							ID:        tc.OfFunction.ID,
							Name:      tc.OfFunction.Function.Name,
							Arguments: tc.OfFunction.Function.Arguments,
						})
					}
				}
				result = append(result, rv)
			}

		case m.OfTool != nil:
			t := m.OfTool
			result = append(result, vo.RoundMessageVO{
				Role:    "tool",
				ToolID:  t.ToolCallID,
				Content: t.Content.OfString.Value,
			})
		}
	}
	return result
}

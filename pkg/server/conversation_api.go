package server

import (
	"github.com/gin-gonic/gin"
	"github.com/tianniu-ai/tianniu/pkg/vo"
)

// POST /conversation
func (s *Server) createConversation(c *gin.Context) {
	var req vo.CreateConversationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, StatusInvalidParam, err)
		return
	}
	req.UserID = c.MustGet("userID").(string)

	result, err := s.svc.CreateConversation(req)
	if err != nil {
		respondError(c, StatusInternalServerError, err)
		return
	}

	respondSuccess(c, result)
}

// GET /conversation
func (s *Server) listConversations(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	result, err := s.svc.ListConversations(userID)
	if err != nil {
		respondError(c, StatusInternalServerError, err)
		return
	}

	respondSuccess(c, result)
}

// PATCH /conversation/:conversation_id
func (s *Server) renameConversation(c *gin.Context) {
	conversationID := c.Param("conversation_id")

	var req vo.UpdateConversationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, StatusInvalidParam, err)
		return
	}

	userId := c.MustGet("userID").(string)

	result, err := s.svc.RenameConversation(userId, conversationID, req.Title)
	if err != nil {
		respondError(c, StatusInternalServerError, err)
		return
	}

	respondSuccess(c, result)
}

// DELETE /conversation/:conversation_id
func (s *Server) deleteConversation(c *gin.Context) {
	conversationID := c.Param("conversation_id")
	userId := c.MustGet("userID").(string)

	if err := s.svc.DeleteConversation(userId, conversationID); err != nil {
		respondError(c, StatusInternalServerError, err)
		return
	}

	respondSuccess(c, map[string]any{"conversation_id": conversationID})
}

package server

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/tianniu-ai/tianniu/pkg/agent"
	"github.com/tianniu-ai/tianniu/pkg/vo"
)

// GET /conversation/:conversation_id/message
func (s *Server) listMessages(c *gin.Context) {
	conversationID := c.Param("conversation_id")
	userID := c.MustGet("userID").(string)

	result, err := s.svc.ListMessages(userID, conversationID)
	if err != nil {
		respondError(c, StatusInternalServerError, err)
		return
	}

	respondSuccess(c, result)
}

// POST /conversation/:conversation_id/message
// Create new message and stream agent response via SSE
func (s *Server) createMessage(c *gin.Context) {
	conversationID := c.Param("conversation_id")

	var req vo.CreateMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, StatusInvalidParam, err)
		return
	}

	userId := c.MustGet("userID").(string)
	req.UserID = userId

	eventCh := make(chan vo.SSEMessageVO, 64)
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("panic in createMessage goroutine: %v\n%s", r, debug.Stack())
			}
			close(eventCh)
		}()
		if err := s.svc.CreateMessage(c.Request.Context(), conversationID, req, eventCh); err != nil {
			errMsg := err.Error()
			// Use non-blocking send to avoid deadlock when client has disconnected
			select {
			case eventCh <- vo.SSEMessageVO{Event: agent.EventError, Content: &errMsg}:
			default:
			}
			return
		}
	}()

	for {
		select {
		case <-c.Request.Context().Done():
			log.Warn("client disconnected, aborting SSE stream")
			return
		case e, ok := <-eventCh:
			if !ok {
				return
			}
			c.SSEvent("message", e)
			c.Writer.Flush()
		}
	}
}

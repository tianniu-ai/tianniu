package server

import (
	"github.com/gin-gonic/gin"
	"github.com/liyue201/tian-niu/pkg/agent"
	"github.com/liyue201/tian-niu/pkg/shared/log"
	"github.com/liyue201/tian-niu/pkg/vo"
)

// GET /conversation/:conversation_id/message
func (s *Server) listMessages(c *gin.Context) {
	conversationID := c.Param("conversation_id")

	result, err := s.svc.ListMessages(conversationID)
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
		defer close(eventCh)
		if err := s.svc.CreateMessage(c.Request.Context(), conversationID, req, eventCh); err != nil {
			errMsg := err.Error()
			eventCh <- vo.SSEMessageVO{Event: agent.EventError, Content: &errMsg}
			return
		}
	}()

	for {
		select {
		case <-c.Request.Context().Done():
			log.Warn("Server is shutting down. Exiting...")
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

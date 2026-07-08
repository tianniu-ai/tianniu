package server

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/tianniu-ai/tianniu/pkg/repository"
	"github.com/tianniu-ai/tianniu/pkg/vo"
)

// POST /user/register
func (s *Server) register(c *gin.Context) {
	var req vo.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, StatusInvalidParam, err)
		return
	}

	res, err := s.svc.Register(req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEntry) {
			respondError(c, StatusDuplicateEntry, err)
			return
		}
		respondError(c, StatusInternalServerError, err)
		return
	}

	respondSuccess(c, res)
}

// POST /user/login
func (s *Server) login(c *gin.Context) {
	var req vo.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, StatusInvalidParam, err)
		return
	}

	res, err := s.svc.Login(req)
	if err != nil {
		respondError(c, StatusInternalServerError, err)
		return
	}

	respondSuccess(c, res)
}

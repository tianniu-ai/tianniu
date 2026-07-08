package server

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
	"github.com/tianniu-ai/tianniu/pkg/agent"
	"github.com/tianniu-ai/tianniu/pkg/auth"
	"github.com/tianniu-ai/tianniu/pkg/repository"
	"github.com/tianniu-ai/tianniu/pkg/service"
)

type Server struct {
	svc        *service.Service
	httpServer *http.Server
	wg         sync.WaitGroup
}

func NewServer(addr string, db *repository.Repository, mgr *agent.Manager) *Server {
	svc := service.NewService(db, mgr)
	engine := gin.New()
	gin.SetMode(gin.ReleaseMode)
	engine.Use(gin.Recovery(), gin.Logger())

	s := &Server{
		svc:        svc,
		httpServer: &http.Server{Addr: addr, Handler: engine},
	}
	s.setupRouter(engine)
	return s
}

func (s *Server) setupRouter(g *gin.Engine) {
	s.setCors(g)

	api := g.Group("/api")

	// Public routes (no authentication required)
	api.POST("/user/register", s.register)
	api.POST("/user/login", s.login)

	// Protected routes (require JWT authentication)
	protected := api.Group("/")
	protected.Use(jwtMiddleware())
	protected.POST("/conversation", s.createConversation)
	protected.GET("/conversation", s.listConversations)
	protected.PATCH("/conversation/:conversation_id", s.renameConversation)
	protected.DELETE("/conversation/:conversation_id", s.deleteConversation)
	protected.POST("/conversation/:conversation_id/message", s.createMessage)
	protected.GET("/conversation/:conversation_id/message", s.listMessages)
}

// jwtMiddleware is a Gin middleware to validate JWT tokens
func jwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Authorization header is required"})
			c.Abort()
			return
		}

		// Bearer token format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info to context
		c.Set("userID", claims["user_id"])
		c.Set("username", claims["username"])
		c.Next()
	}
}

func (s *Server) Run() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := s.httpServer.ListenAndServe()
		if err != nil {
			log.Infof("%v", err.Error())
		}
	}()
}

func (s *Server) Stop() {
	s.httpServer.Shutdown(context.Background())
	s.wg.Wait()
}

func (s *Server) setCors(r gin.IRouter) {
	corsCfg := cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Access-Control-Allow-Origin", "Accept",
			"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsCfg))
}

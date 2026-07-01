package vo

// R is the unified JSON response wrapper
type R struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// CreateConversationReq POST /conversation request body
type CreateConversationReq struct {
	UserID string `json:"user_id"`
	Title  string `json:"title"  binding:"required"`
}

// UpdateConversationReq PATCH /conversation/{id} request body
type UpdateConversationReq struct {
	Title string `json:"title" binding:"required"`
}

// CreateMessageReq POST /conversation/{id}/message request body
type CreateMessageReq struct {
	UserID          string `json:"user_id""`
	Query           string `json:"query" binding:"required"`
	ParentMessageID string `json:"parent_message_id"`
}

// ConversationVO GET /conversation list item
type ConversationVO struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Title     string `json:"title"`
	CreatedAt int64  `json:"created_at"`
}

// RoundMessageVO is a simplified view of an LLM round message
type RoundMessageVO struct {
	Role      string       `json:"role"`                 // user / assistant / tool
	Content   string       `json:"content,omitempty"`    // text content
	ToolCalls []ToolCallVO `json:"tool_calls,omitempty"` // tool calls initiated by assistant
	ToolName  string       `json:"tool_name,omitempty"`  // tool message's tool name
	ToolID    string       `json:"tool_id,omitempty"`    // call_id corresponding to tool message
}

// ToolCallVO is a simplified view of a tool call
type ToolCallVO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatMessageVO GET /conversation/{id}/message list item
type ChatMessageVO struct {
	MessageID       string           `json:"id"`
	ConversationID  string           `json:"conversation_id"`
	ParentMessageID string           `json:"parent_message_id"`
	Query           string           `json:"query"`
	Response        string           `json:"response"`
	Model           string           `json:"model"`
	CreatedAt       int64            `json:"created_at"`
	Rounds          []RoundMessageVO `json:"rounds,omitempty"`
}

// RegisterReq POST /user/register request body
type RegisterReq struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required"`
}

// LoginReq POST /user/login request body
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserVO User info response
type UserVO struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt int64  `json:"created_at"`
}

// LoginRespVO Login response
type LoginRespVO struct {
	User  UserVO `json:"user"`
	Token string `json:"token"`
}

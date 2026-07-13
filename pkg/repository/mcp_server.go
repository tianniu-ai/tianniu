package repository

import (
	"errors"

	"github.com/tianniu-ai/tianniu/pkg/model"
	"gorm.io/gorm"
)

// GetAllMcpServers returns all MCP servers
func (r *SQLStore) GetAllMcpServers() ([]*model.McpServer, error) {
	var servers []model.McpServer
	err := r.db.Find(&servers).Error
	if err != nil {
		return nil, err
	}
	result := make([]*model.McpServer, 0, len(servers))
	for i := range servers {
		result = append(result, &servers[i])
	}
	return result, nil
}

// GetMcpServerByID returns a MCP server by ID
func (r *SQLStore) GetMcpServerByID(id string) (*model.McpServer, error) {
	var s model.McpServer
	err := r.db.Where("id = ?", id).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("mcp server not found")
		}
		return nil, err
	}
	return &s, nil
}

// GetMcpServerByName returns a system MCP server by name
func (r *SQLStore) GetMcpServerByName(name string) (*model.McpServer, error) {
	var s model.McpServer
	err := r.db.Where("name = ? AND user_id = ?", name, "").First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("mcp server not found")
		}
		return nil, err
	}
	return &s, nil
}

// GetMcpServersByUserID returns MCP servers by user ID
func (r *SQLStore) GetMcpServersByUserID(userID string) ([]*model.McpServer, error) {
	var servers []model.McpServer
	err := r.db.Where("user_id = ?", userID).Find(&servers).Error
	if err != nil {
		return nil, err
	}
	result := make([]*model.McpServer, 0, len(servers))
	for i := range servers {
		result = append(result, &servers[i])
	}
	return result, nil
}

// GetSystemMcpServers returns all system MCP servers
func (r *SQLStore) GetSystemMcpServers() ([]*model.McpServer, error) {
	var servers []model.McpServer
	err := r.db.Where("type = ?", "system").Find(&servers).Error
	if err != nil {
		return nil, err
	}
	result := make([]*model.McpServer, 0, len(servers))
	for i := range servers {
		result = append(result, &servers[i])
	}
	return result, nil
}

// GetUserMcpServers returns user-specific MCP servers
func (r *SQLStore) GetUserMcpServers(userID string) ([]*model.McpServer, error) {
	var servers []model.McpServer
	err := r.db.Where("type = ? AND user_id = ?", "user", userID).Find(&servers).Error
	if err != nil {
		return nil, err
	}
	result := make([]*model.McpServer, 0, len(servers))
	for i := range servers {
		result = append(result, &servers[i])
	}
	return result, nil
}

// GetMcpServerForUser returns a MCP server for a specific user
func (r *SQLStore) GetMcpServerForUser(userID, serverName string) (*model.McpServer, error) {
	var s model.McpServer
	err := r.db.Where("name = ? AND (user_id = ? OR user_id = ?)", serverName, userID, "").First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("mcp server not found")
		}
		return nil, err
	}
	return &s, nil
}

// SaveMcpServer saves a MCP server to the database
func (r *SQLStore) SaveMcpServer(s *model.McpServer) error {
	return r.db.Save(s).Error
}

// DeleteMcpServer deletes a MCP server by ID
func (r *SQLStore) DeleteMcpServer(id string) error {
	result := r.db.Where("id = ?", id).Delete(&model.McpServer{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("mcp server not found")
	}
	return nil
}

// UpdateMcpServerStatus updates a MCP server's status
func (r *SQLStore) UpdateMcpServerStatus(id string, status string) error {
	result := r.db.Model(&model.McpServer{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("mcp server not found")
	}
	return nil
}

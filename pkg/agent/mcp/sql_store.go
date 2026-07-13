package mcp

import (
	"github.com/tianniu-ai/tianniu/pkg/repository"
)

type SQLMcpStore struct {
	repo *repository.SQLStore
}

func NewSQLMcpStore(repo *repository.SQLStore) *SQLMcpStore {
	return &SQLMcpStore{repo: repo}
}

func (s *SQLMcpStore) GetAll() ([]*McpServer, error) {
	modelServers, err := s.repo.GetAllMcpServers()
	if err != nil {
		return nil, err
	}
	return convertModelMcpServersToMcpServer(modelServers), nil
}

func (s *SQLMcpStore) GetByID(id string) (*McpServer, error) {
	modelServer, err := s.repo.GetMcpServerByID(id)
	if err != nil {
		return nil, err
	}
	return convertModelToMcpServer(modelServer), nil
}

func (s *SQLMcpStore) GetByName(name string) (*McpServer, error) {
	modelServer, err := s.repo.GetMcpServerByName(name)
	if err != nil {
		return nil, err
	}
	return convertModelToMcpServer(modelServer), nil
}

func (s *SQLMcpStore) GetByUserID(userID string) ([]*McpServer, error) {
	modelServers, err := s.repo.GetMcpServersByUserID(userID)
	if err != nil {
		return nil, err
	}
	return convertModelMcpServersToMcpServer(modelServers), nil
}

func (s *SQLMcpStore) GetSystemMcpServers() ([]*McpServer, error) {
	modelServers, err := s.repo.GetSystemMcpServers()
	if err != nil {
		return nil, err
	}
	return convertModelMcpServersToMcpServer(modelServers), nil
}

func (s *SQLMcpStore) GetUserMcpServers(userID string) ([]*McpServer, error) {
	modelServers, err := s.repo.GetUserMcpServers(userID)
	if err != nil {
		return nil, err
	}
	return convertModelMcpServersToMcpServer(modelServers), nil
}

func (s *SQLMcpStore) GetMcpServerForUser(userID, serverName string) (*McpServer, error) {
	modelServer, err := s.repo.GetMcpServerForUser(userID, serverName)
	if err != nil {
		return nil, err
	}
	return convertModelToMcpServer(modelServer), nil
}

func (s *SQLMcpStore) Save(server *McpServer) error {
	modelServer := convertMcpServerToModel(server)
	return s.repo.SaveMcpServer(modelServer)
}

func (s *SQLMcpStore) Delete(id string) error {
	return s.repo.DeleteMcpServer(id)
}

func (s *SQLMcpStore) UpdateStatus(id string, status McpStatus) error {
	return s.repo.UpdateMcpServerStatus(id, string(status))
}

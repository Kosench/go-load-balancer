package client

import (
	"errors"
	"sync"
	"time"
)

type ClientStore interface {
	Create(client *Client) error
	Get(clientID string) (*Client, error)
	Update(client *Client) error
	Delete(clientID string) error
	List() ([]*Client, error)
}

type InMemoryClientStore struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func NewInMemoryClientStore() *InMemoryClientStore {
	return &InMemoryClientStore{
		clients: make(map[string]*Client),
	}
}

func (s *InMemoryClientStore) Create(client *Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exist := s.clients[client.ID]; exist {
		return errors.New("client already exist")
	}
	if client.CreatedAt.IsZero() {
		client.CreatedAt = time.Now()
	}
	s.clients[client.ID] = client
	return nil
}

func (s *InMemoryClientStore) Get(clientID string) (*Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, exist := s.clients[clientID]
	if !exist {
		return nil, errors.New("client not found")
	}
	return c, nil
}

func (s *InMemoryClientStore) Update(client *Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.clients[client.ID]
	if !exists {
		return errors.New("client not found")
	}
	s.clients[client.ID] = client
	return nil
}

func (s *InMemoryClientStore) Delete(clientID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exist := s.clients[clientID]
	if !exist {
		return errors.New("client not found")
	}
	delete(s.clients, clientID)
	return nil
}

func (s *InMemoryClientStore) List() ([]*Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]*Client, 0, len(s.clients))
	for _, c := range s.clients {
		result = append(result, c)
	}
	return result, nil
}

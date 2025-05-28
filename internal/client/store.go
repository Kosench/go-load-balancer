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

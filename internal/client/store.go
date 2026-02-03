package client

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrClientAlreadyExists is returned when attempting to create a client with an ID or API key that already exists.
	ErrClientAlreadyExists = errors.New("client already exists")
	// ErrClientNotFound is returned when a requested client cannot be found.
	ErrClientNotFound = errors.New("client not found")
	// ErrInvalidClient is returned when client validation fails.
	ErrInvalidClient = errors.New("invalid client")
)

// ClientStore defines the interface for client data persistence and retrieval.
// Implementations must be thread-safe.
type ClientStore interface {
	Create(client *Client) error                // Create a new client
	Get(clientID string) (*Client, error)       // Get client by ID
	GetByAPIKey(apiKey string) (*Client, error) // Get client by API key
	Update(client *Client) error                // Update existing client
	Delete(clientID string) error               // Delete client by ID
	List() ([]*Client, error)                   // List all clients
}

// InMemoryClientStore is a thread-safe in-memory implementation of ClientStore.
// It maintains indexes by both client ID and API key for efficient lookups.
type InMemoryClientStore struct {
	mu          sync.RWMutex
	clients     map[string]*Client // Index by client ID
	apiKeyIndex map[string]*Client // Index by API key
}

// NewInMemoryClientStore creates a new in-memory client store.
func NewInMemoryClientStore() *InMemoryClientStore {
	return &InMemoryClientStore{
		clients:     make(map[string]*Client),
		apiKeyIndex: make(map[string]*Client),
	}
}

func (s *InMemoryClientStore) Create(client *Client) error {
	if err := validateClient(client); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[client.ID]; exists {
		return ErrClientAlreadyExists
	}

	if _, exists := s.apiKeyIndex[client.APIKey]; exists {
		return fmt.Errorf("%w: API key already in use", ErrClientAlreadyExists)
	}

	if client.CreatedAt.IsZero() {
		client.CreatedAt = time.Now()
	}

	s.clients[client.ID] = client
	s.apiKeyIndex[client.APIKey] = client
	return nil
}

func (s *InMemoryClientStore) Get(clientID string) (*Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, exists := s.clients[clientID]
	if !exists {
		return nil, ErrClientNotFound
	}
	return c, nil
}

func (s *InMemoryClientStore) GetByAPIKey(apiKey string) (*Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, exists := s.apiKeyIndex[apiKey]
	if !exists {
		return nil, ErrClientNotFound
	}
	return c, nil
}

func (s *InMemoryClientStore) Update(client *Client) error {
	if err := validateClient(client); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldClient, exists := s.clients[client.ID]
	if !exists {
		return ErrClientNotFound
	}

	// If API key changed, update the index
	if oldClient.APIKey != client.APIKey {
		if _, exists := s.apiKeyIndex[client.APIKey]; exists {
			return fmt.Errorf("%w: API key already in use", ErrClientAlreadyExists)
		}
		delete(s.apiKeyIndex, oldClient.APIKey)
		s.apiKeyIndex[client.APIKey] = client
	}

	s.clients[client.ID] = client
	return nil
}

func (s *InMemoryClientStore) Delete(clientID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, exists := s.clients[clientID]
	if !exists {
		return ErrClientNotFound
	}

	delete(s.clients, clientID)
	delete(s.apiKeyIndex, client.APIKey)
	return nil
}

func (s *InMemoryClientStore) List() ([]*Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Client, 0, len(s.clients))
	for _, c := range s.clients {
		result = append(result, c)
	}
	return result, nil
}

func validateClient(client *Client) error {
	if client == nil {
		return fmt.Errorf("%w: client is nil", ErrInvalidClient)
	}
	if client.ID == "" {
		return fmt.Errorf("%w: client ID cannot be empty", ErrInvalidClient)
	}
	if client.APIKey == "" {
		return fmt.Errorf("%w: API key cannot be empty", ErrInvalidClient)
	}
	if client.Capacity <= 0 {
		return fmt.Errorf("%w: capacity must be greater than 0", ErrInvalidClient)
	}
	if client.RatePerSec <= 0 {
		return fmt.Errorf("%w: rate per second must be greater than 0", ErrInvalidClient)
	}
	return nil
}

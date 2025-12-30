package client

import (
	"sync"
	"time"

	"github.com/SSHcom/privx-sdk-go/v2/restapi"
)

// ConnectionPool manages PrivX API connections to avoid authentication races
type ConnectionPool struct {
	mu        sync.RWMutex
	connector *restapi.Connector
	lastAuth  time.Time
	config    ConnectionConfig
}

type ConnectionConfig struct {
	APIBaseURL        string
	BearerToken       string
	APIClientID       string
	APIClientSecret   string
	OAuthClientID     string
	OAuthClientSecret string
}

var (
	globalPool *ConnectionPool
	poolMu     sync.Mutex
)

// GetConnector returns a shared connector instance
func GetConnector(config ConnectionConfig) (*restapi.Connector, error) {
	poolMu.Lock()
	defer poolMu.Unlock()

	// Initialize global pool if needed
	if globalPool == nil {
		globalPool = &ConnectionPool{
			config: config,
		}
	}

	return globalPool.getOrCreateConnector()
}

func (p *ConnectionPool) getOrCreateConnector() (*restapi.Connector, error) {
	p.mu.RLock()

	// Return existing connector if it's recent (less than 5 minutes old)
	if p.connector != nil && time.Since(p.lastAuth) < 5*time.Minute {
		defer p.mu.RUnlock()
		return p.connector, nil
	}
	p.mu.RUnlock()

	// Need to create new connector
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if p.connector != nil && time.Since(p.lastAuth) < 5*time.Minute {
		return p.connector, nil
	}

	// Create new connector
	connector, err := NewConnector(
		p.config.APIBaseURL,
		p.config.BearerToken,
		p.config.APIClientID,
		p.config.APIClientSecret,
		p.config.OAuthClientID,
		p.config.OAuthClientSecret,
	)

	if err != nil {
		return nil, err
	}

	p.connector = connector
	p.lastAuth = time.Now()

	return connector, nil
}

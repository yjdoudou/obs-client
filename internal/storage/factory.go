package storage

import (
	"fmt"

	"obs-client/internal/connection"
)

type ProviderFactory func(*connection.Connection) (StorageProvider, error)

var providerFactories = map[string]ProviderFactory{}

func RegisterProvider(providerType string, factory ProviderFactory) {
	providerFactories[providerType] = factory
}

func NewProvider(conn *connection.Connection) (StorageProvider, error) {
	factory, ok := providerFactories[conn.Provider]
	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", conn.Provider)
	}
	return factory(conn)
}

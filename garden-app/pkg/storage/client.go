package storage

import (
	"fmt"

	"github.com/calvinmclean/automated-garden/garden-app/pkg"
	"github.com/calvinmclean/automated-garden/garden-app/pkg/storage/configmap"
	"github.com/calvinmclean/automated-garden/garden-app/pkg/storage/yaml"
	"github.com/calvinmclean/automated-garden/garden-app/pkg/weather"
	"github.com/rs/xid"
)

// Config is used to identify and configure a storage client
type Config struct {
	Type    string            `mapstructure:"type"`
	Options map[string]string `mapstructure:"options"`
}

// Client is a "generic" interface used to interact with our storage backend (DB, file, etc)
type Client interface {
	GetGarden(xid.ID) (*pkg.Garden, error)
	GetGardens(bool) ([]*pkg.Garden, error)
	SaveGarden(*pkg.Garden) error
	DeleteGarden(xid.ID) error

	GetZone(xid.ID, xid.ID) (*pkg.Zone, error)
	GetZones(xid.ID, bool) ([]*pkg.Zone, error)
	SaveZone(xid.ID, *pkg.Zone) error
	DeleteZone(xid.ID, xid.ID) error

	GetPlant(xid.ID, xid.ID) (*pkg.Plant, error)
	GetPlants(xid.ID, bool) ([]*pkg.Plant, error)
	SavePlant(xid.ID, *pkg.Plant) error
	DeletePlant(xid.ID, xid.ID) error

	GetWeatherClient(xid.ID) (*weather.Config, error)
	GetWeatherClients(bool) ([]*weather.Config, error)
	SaveWeatherClient(*weather.Config) error
	DeleteWeatherClient(xid.ID) error
}

// NewClient will use the config to create and return the correct type of storage client
func NewClient(config Config) (Client, error) {
	switch config.Type {
	case "YAML", "yaml":
		return yaml.NewClient(config.Options)
	case "ConfigMap", "configmap":
		return configmap.NewClient(config.Options)
	default:
		return nil, fmt.Errorf("invalid type '%s'", config.Type)
	}
}

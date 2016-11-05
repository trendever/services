package gateway

import (
	"payments/models"
)

// Gateways contains all registered gw after LoadAll is called
var Gateways = map[string]Gateway{}

// Loaders is a place where you place your own loader
var Loaders []Loader

// Gateway interface (1-step payment)
type Gateway interface {

	// create buying session
	Buy(sess *models.Payment, ipAddr string) (*models.Session, error)

	// get redirect URL for this session
	Redirect(*models.Session) string

	CheckStatus(*models.Session) (finished bool, err error)

	GatewayType() string
}

// Loader is loader (config2gateway)
type Loader interface {
	Load() (enabled bool, gw Gateway, err error)
}

// LoadAll calls all loaders to get all gateways
func LoadAll() error {

	for _, loader := range Loaders {

		enabled, gateway, err := loader.Load()
		if err != nil {
			return err
		} else if !enabled {
			continue
		}

		Gateways[gateway.GatewayType()] = gateway
	}

	return nil
}

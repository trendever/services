package gateway

import (
	"payments/models"
	"proto/payment"
)

// Gateways contains all registered gw after LoadAll is called
var Gateways = map[string]Gateway{}

// Loaders is a place where you place your own loader
var Loaders []Loader

// Gateway interface (1-step payment)
type Gateway interface {
	GatewayType() string
}

// PaymentGateway payment functions
type PaymentGateway interface {
	// create buying session
	Buy(sess *models.Payment, info *payment.UserInfo) (*models.Session, error)

	// get redirect URL for this session
	Redirect(*models.Session) string

	CheckStatus(*models.Session) (finished bool, err error)
}

// CardGateway adittional gw type
type CardGateway interface {
	Add(info *payment.UserInfo) (string, error)
	GetCards(info *payment.UserInfo) ([]*payment.Card, error)
}

// Loader is loader (config2gateway)
type Loader interface {
	Load() (gw []Gateway, err error)
}

// LoadAll calls all loaders to get all gateways
func LoadAll() error {

	for _, loader := range Loaders {

		gws, err := loader.Load()
		if err != nil {
			return err
		}

		for _, gw := range gws {
			Gateways[gw.GatewayType()] = gw
		}
	}

	return nil
}

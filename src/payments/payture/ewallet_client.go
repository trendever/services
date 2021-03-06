package payture

import (
	"errors"

	"payments/config"
	"payments/gateway"
)

const vwGwType = "payture_ewallet"

// Ewallet def
type Ewallet struct {
	config.Ewallet
}

// EwalletLoader loads this gw
type EwalletLoader struct {
}

// GetSandboxEwallet returns testing client
func GetSandboxEwallet() *Ewallet {

	return &Ewallet{
		Ewallet: config.Ewallet{
			URL:      "https://sandbox2.payture.com",
			KeyAdd:   "VWMerchantTrendeverAdd",
			KeyPay:   "VWMerchantTrendeverPay",
			Password: "123",
		},
	}

}

// GetEwallet returns payture service client
func GetEwallet() *Ewallet {

	cfg := config.Get()

	if cfg.Ewallet.Sandbox {
		client := GetSandboxEwallet()
		client.Secret = cfg.Ewallet.Secret
		return client
	}

	return &Ewallet{
		Ewallet: cfg.Ewallet,
	}
}

// Load gateway from config
func (cl *EwalletLoader) Load() (gw []gateway.Gateway, err error) {
	client := GetEwallet()
	if client.URL == "" || client.KeyAdd == "" || client.KeyPay == "" || client.Password == "" || client.Secret == "" {
		return nil, errors.New("Incorrect config")
	}

	return []gateway.Gateway{
		client,
	}, nil
}

// GatewayType for this pkg
func (c *Ewallet) GatewayType() string {
	return vwGwType
}

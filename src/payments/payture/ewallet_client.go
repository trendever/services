package payture

import (
	"payments/config"
	"payments/gateway"
)

// Ewallet def
type Ewallet struct {
	config.Ewallet
}

// EwalletAdd part
type EwalletAdd struct {
	Ewallet
}

// EwalletPay part
type EwalletPay struct {
	Ewallet
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
		return GetSandboxEwallet()
	}

	return &Ewallet{
		Ewallet: cfg.Ewallet,
	}
}

// Load gateway from config
func (cl *EwalletLoader) Load() (gw []gateway.Gateway, err error) {
	client := GetEwallet()
	if client.URL == "" || client.KeyAdd == "" || client.KeyPay == "" || client.Password == "" || client.Secret == "" {
		return nil, nil
	}

	return []gateway.Gateway{
		&EwalletAdd{*client},
		&EwalletPay{*client},
	}, nil
}

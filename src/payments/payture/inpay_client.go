package payture

import (
	"payments/config"
	"payments/gateway"
)

// InPay def
type InPay struct {
	URL string
	Key string
}

// Loader loads this gw
type Loader struct {
}

func init() {
	gateway.Loaders = append(gateway.Loaders, &Loader{})
	gateway.Loaders = append(gateway.Loaders, &EwalletLoader{})
}

// GetSandboxClient returns testing client
func GetSandboxClient() *InPay {
	return &InPay{
		URL: "https://sandbox2.payture.com",
		Key: "Merchant",
	}
}

// GetClient returns payture service client
func GetClient() *InPay {

	if config.Get().Payture.Sandbox {
		return GetSandboxClient()
	}

	return &InPay{
		URL: config.Get().Payture.URL,
		Key: config.Get().Payture.Key,
	}
}

// Load gateway from config
func (cl *Loader) Load() (gw []gateway.Gateway, err error) {
	client := GetClient()
	if client.URL == "" || client.Key == "" {
		return nil, nil
	}

	return []gateway.Gateway{client}, nil
}

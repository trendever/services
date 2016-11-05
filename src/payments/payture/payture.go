package payture

import (
	"payments/config"
	"payments/gateway"
)

// Client def
type Client struct {
	URL string
	Key string
}

// Loader loads this gw
type Loader struct {
}

func init() {
	gateway.Loaders = append(gateway.Loaders, &Loader{})
}

// GetSandboxClient returns testing client
func GetSandboxClient() *Client {

	return &Client{
		URL: "https://sandbox2.payture.com",
		Key: "Merchant",
	}

}

// GetClient returns payture service client
func GetClient() *Client {

	if config.Get().Payture.Sandbox {
		return GetSandboxClient()
	}

	return &Client{
		URL: config.Get().Payture.URL,
		Key: config.Get().Payture.Key,
	}
}

// Load gateway from config
func (cl *Loader) Load() (enabled bool, gw gateway.Gateway, err error) {
	client := GetClient()
	if client.URL == "" || client.Key == "" {
		return false, nil, nil
	}

	return true, client, nil
}

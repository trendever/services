package payture

import (
	"payments/config"
)

// Client def
type Client struct {
	URL string
	Key string
}

// GetSandboxClient returns testing client
func GetSandboxClient() *Client {

	return &Client{
		URL: "https://sandbox.payture.com",
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

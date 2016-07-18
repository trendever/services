package bitly

import (
	"github.com/timehop/go-bitly"
	"utils/log"
	"net/url"
	"auth/config"
)

// GetBitly returns Bitly client
func GetBitly() *bitly.Client {
	settings := config.Get().Bitly
	return &bitly.Client{APIKey: settings.ApiKey, Login: settings.Login, AccessToken: settings.AccessToken}
}

func GetSiteUrl(token string) string {
	v := &url.Values{}
	v.Add("token", token)
	u, err := url.Parse(config.Get().SiteUrl)
	if err != nil {
		log.Error(err)
		return ""
	}
	u.RawQuery = v.Encode()
	return u.String()
}

func GetShortUrl(url string) (bitly.ShortenResult, error) {
	return GetBitly().Shorten(url)
}

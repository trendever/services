package bitly

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var (
	apiBase = "https://api-ssl.bitly.com"
)

type Client struct {
	// Authentication
	APIKey string

	Login       string
	AccessToken string
}

type response struct {
	StatusCode int    `json:"status_code"`
	StatusTxt  string `json:"status_txt"`
}

type ShortenResult struct {
	GlobalHash string `json:"global_hash"`
	Hash       string `json:"hash"`
	LongURL    string `json:"long_url"`
	NewHash    int    `json:"new_hash"`
	URL        string `json:"url"`
}

type errorResponse struct {
	Data []string
	response
}

type shortenResponse struct {
	Data ShortenResult
	response
}

func (c Client) valid() bool {
	if c.AccessToken != "" {
		return true
	}

	if c.Login != "" && c.APIKey != "" {
		return true
	}

	return false
}

func (c Client) Shorten(long string) (ShortenResult, error) {
	return c.ShortenUsingDomain(long, "")
}

func (c Client) ShortenUsingDomain(long string, domain string) (ShortenResult, error) {
	var resp shortenResponse

	params := url.Values{"longUrl": {long}}
	if domain != "" {
		params.Set("domain", domain)
	}

	err := c.runRequestAndMarshal("GET", apiBase+"/v3/shorten", &params, &resp)
	if err != nil {
		return resp.Data, err
	}

	return resp.Data, nil
}

func (c Client) runRequestAndMarshal(method string, endpoint string, params *url.Values, dest interface{}) error {
	body, err := c.runRequest(method, endpoint, params)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, dest)
	if err != nil {
		var errResp errorResponse
		err = json.Unmarshal(body, &errResp)
		if err != nil {
			return err
		}

		return fmt.Errorf("Bitly Error: %v", errResp.StatusTxt)
	}

	return nil
}

func (c Client) runRequest(method string, endpoint string, params *url.Values) ([]byte, error) {
	var (
		err error
		req *http.Request
		b   io.Reader
	)

	if !c.valid() {
		return []byte{}, fmt.Errorf("Authentication missing")
	}

	if params == nil {
		params = &url.Values{}
	}

	c.addAuthentication(params)

	if method == "GET" {
		enc := params.Encode()
		if enc != "" {
			endpoint = endpoint + "?" + enc
		}
		req, err = http.NewRequest(method, endpoint, nil)
	} else if method == "POST" {
		b = strings.NewReader(params.Encode())
		req, err = http.NewRequest(method, endpoint, b)
	} else {
		err = fmt.Errorf("Method not supported: %v", method)
	}

	if err != nil {
		return []byte{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (c Client) addAuthentication(params *url.Values) {
	if c.Login != "" {
		params.Set("login", c.Login)
	}

	if c.APIKey != "" {
		params.Set("apiKey", c.APIKey)
	}

	if c.AccessToken != "" {
		params.Set("access_token", c.AccessToken)
	}
}

func (r ShortenResult) IsNewHash() bool {
	return r.NewHash == 1
}

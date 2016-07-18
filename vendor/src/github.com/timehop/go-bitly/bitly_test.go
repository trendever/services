package bitly

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	shortenResponseStr = `{
    "data": {
      "global_hash": "900913",
      "hash": "ze6poY",
      "long_url": "http://google.com/",
      "new_hash": 0,
      "url": "http://bit.ly/ze6poY"
    },
    "status_code": 200,
    "status_txt": "OK"
  }`
	shortenResponseAsNewHashStr = `{
    "data": {
      "global_hash": "900913",
      "hash": "ze6poY",
      "long_url": "http://google.com/",
      "new_hash": 1,
      "url": "http://bit.ly/ze6poY"
    },
    "status_code": 200,
    "status_txt": "OK"
  }`
	shortenResponseWithErrStr = `{
    "data": [ ],
    "status_code": 500,
    "status_txt": "MISSING_ARG_ACCESS_TOKEN"
  }`
)

func setupMockServer() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v3/shorten" {
			fmt.Fprintln(w, shortenResponseStr)
		} else if r.URL.Path == "/v3/error" {
			fmt.Fprintln(w, shortenResponseWithErrStr)
		}
	}))
	apiBase = ts.URL
	return ts
}

func TestClient(t *testing.T) {
	var c Client

	c = Client{}
	if c.valid() {
		t.Errorf("Client without access_token or api_key should not be valid")
	}

	c = Client{APIKey: "some_key"}
	if c.valid() {
		t.Errorf("Client with api_key but not login should not be valid")
	}

	c = Client{Login: "login", APIKey: "some_key"}
	if !c.valid() {
		t.Errorf("Client with api_key should be valid")
	}

	c = Client{AccessToken: "some_access_token"}
	if !c.valid() {
		t.Errorf("Client with access_token should be valid")
	}
}

func TestRunRequestAndMarshal(t *testing.T) {
	setupMockServer()
	var resp shortenResponse

	c := Client{AccessToken: "some_access_token"}
	err := c.runRequestAndMarshal("GET", apiBase+"/v3/error", nil, &resp)
	if err == nil {
		t.Errorf("Mock server did not return error")
	}

  if err.Error() != "Bitly Error: MISSING_ARG_ACCESS_TOKEN" {
		t.Errorf("runRequestAndMarshal did not return the response error message")
  }
}

func TestShortenResult(t *testing.T) {
	var resp shortenResponse

	err := json.Unmarshal([]byte(shortenResponseStr), &resp)
	if err != nil {
		t.Errorf("Could not parse shorten result")
	}
	r := resp.Data

	if r.IsNewHash() {
		t.Error("Shorten with a 0 for new_hash is not a new hash")
	}

	err = json.Unmarshal([]byte(shortenResponseAsNewHashStr), &resp)
	if err != nil {
		t.Errorf("Could not parse shorten result")
	}
	r = resp.Data

	if !r.IsNewHash() {
		t.Error("Shorten with a 1 for new_hash is a new hash")
	}
}

func TestShorten(t *testing.T) {
	setupMockServer()

	c := Client{AccessToken: "some_access_token"}
	r, err := c.Shorten("http://google.com/")

	if err != nil {
		t.Errorf("Shorten returned an unexpected error")
	}

	if r.GlobalHash != "900913" {
		t.Errorf("Shorten response did not parse correctly")
	}

	if r.Hash != "ze6poY" {
		t.Errorf("Shorten response did not parse correctly")
	}

	if r.LongURL != "http://google.com/" {
		t.Errorf("Shorten response did not parse correctly")
	}

	if r.IsNewHash() {
		t.Errorf("Shorten response did not parse correctly")
	}

	if r.URL != "http://bit.ly/ze6poY" {
		t.Errorf("Shorten response did not parse correctly")
	}
}

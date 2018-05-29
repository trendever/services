package instagram

import (
	"bytes"
	"common/log"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// Possible network instagram errors
var (
	ErrorCheckpointRequired = errors.New("Checkpoint action needed to proceed")
	ErrorPageNotFound       = errors.New("Page not found")
)

// Dump all request even if account do not require that.
var ForceDebug = false

// DisableJSONIndent disables indenting json in logs
var DisableJSONIndent = true

// @TODO Use this everywhere
type Error struct {
	Code    int
	Message string
	Raw     []byte
}

func (e Error) Error() string {
	return e.Message
}

type PostContent struct {
	Type   string
	Reader io.Reader
}

func getToken(cook []*http.Cookie) (string, error) {
	for _, cookie := range cook {
		if cookie.Name == "csrftoken" {
			return cookie.Value, nil
		}
	}

	return "", fmt.Errorf("Cookie csrftoken not found")
}

// Request for Login method. Needs to get the authorization cookies.
func (ig *Instagram) requestMain(method, endpoint string, body interface{}, login bool) (*http.Response, error) {
	client := &http.Client{
		Transport: ig.transport,
	}

	// fill-in headers
	header := make(http.Header)
	header.Add("User-Agent", UserAgent)
	header.Add("Accept", "*/*")
	header.Add("X-IG-Capabilities", X_IG_Capabilities)
	header.Add("X-IG-Connection-Type", "WIFI")
	header.Add("Accept-Language", "en-US")
	header.Add("Cookie2", "$Version=1")

	var bodyReader io.Reader
	switch body := body.(type) {
	case nil:
		bodyReader = nil
	case *PostContent:
		bodyReader = body.Reader
		header.Add("Content-type", body.Type)
	case string:
		bodyReader = bytes.NewReader([]byte(body))
		header.Add("Content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	default:
		return nil, errors.New("unsupported body type")
	}

	req, err := http.NewRequest(method, URL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header = header

	// add auth token if needed
	if !login {
		for _, cookie := range ig.Cookies {
			req.AddCookie(cookie)
		}
	}

	// send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Request with five attempts re-login. Re-login if getting error 'login_required'.
func (ig *Instagram) tryRequest(method, endpoint string, body interface{}) ([]byte, error) {

	for attempt := 0; attempt < 3; attempt++ {

		resp, err := ig.requestMain(method, endpoint, body, false)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		jsonBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != 200 {

			if location := resp.Header.Get("Location"); location > "" {
				log.Debug("got non-200 status code %v for endpoint %v with redirect to %v", resp.StatusCode, endpoint, location)
			} else {
				log.Debug("got non-200 status code %v for endpoint %v", resp.StatusCode, endpoint)
			}
		}
		if resp.StatusCode == 404 {
			return nil, ErrorPageNotFound
		}

		if ForceDebug || ig.Debug {
			var buf bytes.Buffer
			err := json.Indent(&buf, jsonBody, "  ", "  ")
			if err == nil {
				if DisableJSONIndent {
					log.Debug("Instagram Response %v (%v): %v", resp.Status, endpoint, string(jsonBody))
				} else {
					log.Debug("Instagram Response %v (%v): %v", resp.Status, endpoint, buf.String())
				}
			} else {
				log.Debug("Instagram response indent failed for endpoint  %v: %v, raw: %v", endpoint, err, string(jsonBody))
			}
		}

		var message *Message
		err = json.Unmarshal(jsonBody, &message)
		if err != nil {
			return nil, err
		}

		if message.Status == "fail" {
			if message.Message == "checkpoint_required" {
				ig.LoggedIn = false
				return nil, ErrorCheckpointRequired
			}
			if message.Message != "login_required" {
				return nil, Error{
					Code:    resp.StatusCode,
					Message: message.Message,
					Raw:     jsonBody,
				}
			}
			// relogin
			ig.LoggedIn = false
			err = ig.Login()
			if err != nil {
				return nil, err
			}
			time.Sleep(time.Millisecond * 500)
		} else {
			return jsonBody, nil
		}
	}

	return nil, errors.New("max_attempts")
}

func (ig *Instagram) RequestRaw(endpoint string) (body []byte, err error) {
	return ig.tryRequest("GET", endpoint, "")
}

func (ig *Instagram) Request(method, endpoint string, result interface{}) error {

	body, err := ig.tryRequest(method, endpoint, "")
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, result)
	return err
}

func (ig *Instagram) jsonRequest(endpoint string, params map[string]string, result interface{}) error {

	encoded, err := json.Marshal(params)
	if err != nil {
		return err
	}

	body, err := ig.tryRequest("POST", endpoint, generateSignature([]byte(encoded)))
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, result)
	return err
}

func (ig *Instagram) postRequest(endpoint string, params map[string]string, result interface{}) error {

	vals := url.Values{}
	for k, v := range params {
		vals.Add(k, v)
	}

	body, err := ig.tryRequest("POST", endpoint, vals.Encode())
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, result)
	return err
}

func (ig *Instagram) postContentRequest(endpoint string, content *PostContent, result interface{}) error {
	body, err := ig.tryRequest("POST", endpoint, content)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, result)
	return err
}

func (ig *Instagram) loginRequest(method, endpoint string, body, result interface{}) ([]*http.Cookie, error) {

	resp, err := ig.requestMain(method, endpoint, body, true)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if ForceDebug || ig.Debug {
		log.Debug("Instagram Response %v (%v): %v", resp.Status, endpoint, string(jsonBody))
	}

	err = json.Unmarshal(jsonBody, result)
	if err != nil {
		return nil, err
	}

	return resp.Cookies(), nil
}

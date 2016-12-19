package instagram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	"utils/log"
)

// Possible network instagram errors
var (
	ErrorCheckpointRequired = errors.New("Checkpoint action needed to proceed")
)

// DoResponseLogging enables full json body output
var DoResponseLogging = false

// DisableJSONIndent disables indenting json in logs
var DisableJSONIndent = true

func getToken(cook []*http.Cookie) (string, error) {
	for _, cookie := range cook {
		if cookie.Name == "csrftoken" {
			return cookie.Value, nil
		}
	}

	return "", fmt.Errorf("Cookie csrftoken not found")
}

// Request for Login method. Needs to get the authorization cookies.
func (ig *Instagram) requestMain(method, endpoint string, body io.Reader, login bool) (*http.Response, error) {

	// create request
	client := &http.Client{}
	req, err := http.NewRequest(method, URL+endpoint, body)
	if err != nil {
		return nil, err
	}

	// fill-in headers
	req.Header.Add("User-Agent", UserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("X-IG-Capabilities", "3Q4=")
	req.Header.Add("X-IG-Connection-Type", "WIFI")
	req.Header.Add("Content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Accept-Language", "en-US")
	req.Header.Add("Cookie2", "$Version=1")

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
func (ig *Instagram) tryRequest(method, endpoint, body string) ([]byte, error) {

	for attempt := 0; attempt < 5; attempt++ {

		resp, err := ig.requestMain(method, endpoint, bytes.NewReader([]byte(body)), false)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		jsonBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if DoResponseLogging {
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
			ig.LoggedIn = false
			if message.Message != "login_required" {
				return nil, errors.New(message.Message)
			}
			if message.Message == "checkpoint_required" {
				return nil, ErrorCheckpointRequired
			}
			// relogin
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

func (ig *Instagram) request(method, endpoint string, result interface{}) error {

	body, err := ig.tryRequest(method, endpoint, "")
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

func (ig *Instagram) loginRequest(method, endpoint, body string, result interface{}) ([]*http.Cookie, error) {

	resp, err := ig.requestMain(method, endpoint, bytes.NewReader([]byte(body)), true)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if DoResponseLogging {
		log.Debug("Instagram Response %v (%v): %v", resp.Status, endpoint, string(jsonBody))
	}

	err = json.Unmarshal(jsonBody, result)
	if err != nil {
		return nil, err
	}

	return resp.Cookies(), nil
}

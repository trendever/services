package instagram

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"utils/log"
)

// Emulate WebView browser

// According to https://github.com/huttarichard/instagram-private-api/blob/master/client/v1/web/challenge.js,
//  iphone UA is the best choice to send requests
var userAgent = fmt.Sprintf("Mozilla/5.0 (iPhone; CPU iPhone OS 9_3_3 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Mobile/13G34 Instagram %v (iPhone7,2; iPhone OS 9_3_3; cs_CZ; cs-CZ; scale=2.00; 750x1334)", Version)

// Possible checkpoint methods
const (
	MethodSms   = "sms"
	MethodEmail = "email"
)

func browserRequest(method, addr, referer string, cookies []*http.Cookie, params map[string]string) (string, []*http.Cookie, error) {

	vals := url.Values{}
	for k, v := range params {
		vals.Add(k, v)
	}

	body := vals.Encode()

	client := &http.Client{}
	req, err := http.NewRequest(method, addr, bytes.NewReader([]byte(body)))
	if err != nil {
		return "", nil, err
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// fill-in headers
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Language", "en-US")
	req.Header.Add("Origin", "https://i.instagram.com")
	if referer != "" {
		req.Header.Add("Referer", referer)
	}

	if method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	// send request
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	if DoResponseLogging {
		log.Debug("URL Values: %v", vals.Encode())
		log.Debug("Checkpoint POST result: %v", string(response))
	}

	return string(response), resp.Cookies(), nil
}

// step1: grab cookies and available login methods
func (ig *Instagram) checkpointStep1() ([]string, error) {

	body, cookies, err := browserRequest("GET", ig.CheckpointURL, "", nil, nil)
	if err != nil {
		return nil, err
	}

	var methods []string
	if strings.Contains(body, `<input type="submit" name="sms" class="checkpoint-button-neutral" value="`) {
		methods = append(methods, MethodSms)
	}
	if strings.Contains(body, `<input type="submit" name="email" class="checkpoint-button-neutral" value="`) {
		methods = append(methods, MethodEmail)
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("Could not start checkpoint process")
	}

	ig.CheckpointCookies = cookies

	return methods, nil
}

// step2: send code using given method
func (ig *Instagram) checkpointStep2(method string) error {

	token, err := getToken(ig.CheckpointCookies)
	if err != nil {
		return err
	}

	values := map[string]string{
		"csrfmiddlewaretoken": token,
	}

	switch method {
	case MethodSms:
		values["sms"] = "Verify by SMS"
	case MethodEmail:
		values["email"] = "Verify by Email"
	default:
		return fmt.Errorf("Incorrect method supplied")
	}

	body, cookies, err := browserRequest("POST", ig.CheckpointURL, ig.CheckpointURL, ig.CheckpointCookies, values)
	if err != nil {
		return err
	}

	if !strings.Contains(body, `<input id="id_response_code" inputmode="numeric" name="response_code"`) {
		return fmt.Errorf("Code input form not found")
	}

	ig.CheckpointCookies = cookies
	return nil
}

// step2: submit code
func (ig *Instagram) checkpointSubmit(code string) error {

	token, err := getToken(ig.CheckpointCookies)
	if err != nil {
		return err
	}

	values := map[string]string{
		"csrfmiddlewaretoken": token,
		"response_code":       code,
	}

	body, _, err := browserRequest("POST", ig.CheckpointURL, ig.CheckpointURL, ig.CheckpointCookies, values)
	if err != nil {
		return err
	}

	// @TODO: how to check if everything is ok?
	_ = body

	// ig.checkpointCookies = nil
	return nil
}

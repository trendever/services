package instagram

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"utils/log"

	// use custom-patched http pkg to allow using backslashes in cookies
	fixedhttp "instagram/http"
	http "net/http"
)

// Emulate WebView browser
var checkpointUserAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 9_3_3 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Mobile/13G34 Instagram " +
	Version +
	" (iPhone7,2; iPhone OS 9_3_3; cs_CZ; cs-CZ; scale=2.00; 750x1334)"

// Possible checkpoint methods
const (
	MethodSms   = "sms"
	MethodEmail = "email"
)

func encode(params map[string]string) string {
	vals := url.Values{}
	for k, v := range params {
		vals.Add(k, v)
	}

	return vals.Encode()
}

func browserRequest(method, addr, referer string, cookies []*http.Cookie, body string) (string, []*http.Cookie, error) {

	client := &fixedhttp.Client{}
	req, err := fixedhttp.NewRequest(method, addr, bytes.NewReader([]byte(body)))
	if err != nil {
		return "", nil, err
	}

	for _, cookie := range cookies {
		req.AddCookie((*fixedhttp.Cookie)(cookie))
	}

	// fill-in headers
	for k, v := range map[string]string{
		"User-Agent":                checkpointUserAgent,
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.5",
		"Connection":                "keep-alive",
		"Origin":                    "https://i.instagram.com",
		"Upgrade-Insecure-Requests": "1",
	} {
		req.Header.Add(k, v)
	}

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

	gotCook := resp.Cookies()
	resCook := []*http.Cookie{}
	for _, v := range gotCook {
		resCook = append(resCook, (*http.Cookie)(v))
	}

	if DoResponseLogging {
		log.Debug("URL: %v", addr)
		log.Debug("REQ headers: %v", req.Header)
		log.Debug("REQ params: %v", body)
		log.Debug("RESP headers: %v", resp.Header)
		log.Debug("Checkpoint POST result: %v", string(response))
	}

	return string(response), concatCookies(cookies, resCook), nil
}

// step1: grab cookies and available login methods
func (ig *Instagram) checkpointStep1() ([]string, error) {

	body, cookies, err := browserRequest("GET", ig.CheckpointURL, "", nil, "")
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

	body, cookies, err := browserRequest("POST", ig.CheckpointURL, ig.CheckpointURL, ig.CheckpointCookies, encode(values))
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

	// I wonder if Instagram devs made post parameters order matter INTENTIONALLY? If yes, they are fucken evil geniouses
	params := fmt.Sprintf("response_code=%v&csrfmiddlewaretoken=%v", code, token)

	body, _, err := browserRequest("POST", ig.CheckpointURL, ig.CheckpointURL, ig.CheckpointCookies, params)
	if err != nil {
		return err
	}

	// @TODO: how to check if everything is ok?
	_ = body

	// ig.checkpointCookies = nil
	return nil
}

func concatCookies(oldCook, newCook []*http.Cookie) []*http.Cookie {

	var (
		res    = newCook
		setted = map[string]bool{}
	)

	for _, cook := range newCook {
		setted[cook.Name] = true
	}

	for _, cook := range oldCook {
		if !setted[cook.Name] {
			res = append(res, cook)
		}
	}

	return res
}

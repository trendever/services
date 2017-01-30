package instagram

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"time"
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

func (ig *Instagram) browserRequest(method, addr, referer string, cookies []*http.Cookie, body string) (string, []*http.Cookie, error) {

	client := &fixedhttp.Client{
		Transport: &fixedhttp.Transport{
			Dial: ig.Dial,
		},
		Timeout: 5 * time.Second,
	}
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

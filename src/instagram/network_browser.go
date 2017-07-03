package instagram

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"time"
	"utils/log"

	// use custom-patched http pkg to allow using backslashes in cookies
	"golang.org/x/net/proxy"
	fixedhttp "instagram/http"
	"net"
	http "net/http"
)

// Emulate WebView browser
var checkpointUserAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 9_3_3 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Mobile/13G34 Instagram " +
	Version +
	" (iPhone7,2; iPhone OS 9_3_3; cs_CZ; cs-CZ; scale=2.00; 750x1334)"

// "fixedhttp" version
func transportFromURL_WTFVersion(proxyURL string) (ret *fixedhttp.Transport, err error) {
	// mostly copy of http.DefaultTransport
	ret = &fixedhttp.Transport{
		Proxy: fixedhttp.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if proxyURL == "" {
		return
	}
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}
	switch parsed.Scheme {
	case "http", "https":
		ret.Proxy = fixedhttp.ProxyURL(parsed)
	default:
		// DialContext in x/net/proxy is on review for now
		ret.DialContext = nil

		var dialer proxy.Dialer
		// correctly supports only socks5
		dialer, err = proxy.FromURL(parsed, &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		})
		if err != nil {
			return
		}
		ret.Dial = dialer.Dial
	}
	return
}

func (ig *Instagram) browserRequest(method, addr, referer string, cookies []*http.Cookie, body string) (string, []*http.Cookie, error) {
	transport, err := transportFromURL_WTFVersion(ig.Proxy)
	if err != nil {
		return "", nil, err
	}

	client := &fixedhttp.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}
	req, err := fixedhttp.NewRequest(method, addr, bytes.NewReader([]byte(body)))
	if err != nil {
		return "", nil, err
	}

	for _, cookie := range cookies {
		req.AddCookie((*fixedhttp.Cookie)(cookie))
		if cookie.Name == "csrftoken" {
			req.Header.Add("X-CSRFToken", cookie.Value)
		}
	}

	// fill-in headers
	for k, v := range map[string]string{
		"User-Agent":                checkpointUserAgent,
		"Accept":                    "*/*",
		"Accept-Language":           "en-US,en;q=0.5",
		"Connection":                "keep-alive",
		"Origin":                    "https://i.instagram.com",
		"X-Instagram-AJAX":          "1",
		"X-Requested-With":          "XMLHttpRequest",
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

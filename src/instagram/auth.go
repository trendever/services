package instagram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"utils/log"
)

// Possible auth instagram errors
var (
	ErrorEmptyPassword = errors.New("Login or password is empty")
)

type ChallengeReply struct {
	Type   string
	Status string
}

// Login to Instagram.
func (ig *Instagram) Login() error {

	fetch := fmt.Sprintf("/si/fetch_headers/?challenge_type=signup&guid=%v", generateUUID(false))

	resp, err := ig.requestMain("GET", fetch, nil, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	token, err := getToken(resp.Cookies())
	if err != nil {
		return err
	}

	if ig.Username == "" || ig.password == "" {
		return ErrorEmptyPassword
	}

	// login
	login := &Login{
		DeviceId:          ig.DeviceID,
		PhoneId:           ig.PhoneID,
		Guid:              ig.UUID,
		UserName:          ig.Username,
		Password:          ig.password,
		Csrftoken:         token,
		LoginAttemptCount: "0",
	}

	jsonData, err := json.Marshal(login)
	if err != nil {
		return err
	}

	var loginResp LoginResponse
	cookies, err := ig.loginRequest("POST", "/accounts/login/?", generateSignature(jsonData), &loginResp)
	if err != nil {
		return err
	}

	ig.Cookies = cookies

	if loginResp.Status == "fail" {
		if loginResp.Message.Message == "checkpoint_required" {
			ig.CheckpointURL = loginResp.CheckpointURL
			uid, err := ig.getUidByCheckpointLink()
			if err != nil {
				return err
			}
			ig.UserID = uid
			ig.updateRankToken()
			return ErrorCheckpointRequired
		}
		return errors.New(loginResp.Message.Message)
	}

	ig.UserID = loginResp.LoggedInUser.Pk
	ig.updateRankToken()
	ig.LoggedIn = true

	return nil
}

var uidRegexp = regexp.MustCompile(`https://i.instagram.com/challenge/([0-9]+)/`)

func (ig *Instagram) getUidByCheckpointLink() (uint64, error) {

	res := uidRegexp.FindStringSubmatch(ig.CheckpointURL)
	if len(res) != 2 {
		return 0, fmt.Errorf("Could not find UID for user %v in checkpoint URL (%v), format changed?: %v", ig.Username, ig.CheckpointURL, res)
	}

	return strconv.ParseUint(res[1], 10, 64)
}

func (ig *Instagram) updateRankToken() {
	ig.RankToken = fmt.Sprintf("%d_%v", ig.UserID, ig.UUID)
}

func (ig *Instagram) checkpointRequest(addr string, referer string, cookies []*http.Cookie, payload string) (string, []*http.Cookie, error) {
	client := &http.Client{
		Transport: ig.transport,
	}

	var req *http.Request
	var err error
	if payload == "" {
		req, err = http.NewRequest("GET", addr, nil)
	} else {
		req, err = http.NewRequest("POST", addr, bytes.NewReader([]byte(payload)))
	}
	if err != nil {
		return "", nil, err
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
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

	if payload != "" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	// send request
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	if DoResponseLogging {
		log.Debug("URL: %v", addr)
		log.Debug("REQ headers: %v", req.Header)
		log.Debug("REQ params: %v", payload)
		log.Debug("RESP headers: %v", resp.Header)
		log.Debug("Checkpoint POST result: %v", string(response))
	}

	return string(response), concatCookies(cookies, resp.Cookies()), nil
}

// SendCode sends checkpoint code
func (ig *Instagram) SendCode(preferEmail bool) (string, error) {

	if ig.CheckpointURL == "" {
		return "", errors.New("Can not send code! Checkpoint URL is empty")
	}

	methods, err := ig.checkpointStep1()
	if err != nil {
		return "", err
	}

	var useMethod string
	for _, method := range methods {
		switch {
		case
			useMethod == "",
			method == MethodEmail && preferEmail,
			method == MethodSms && !preferEmail:

			/*_*/
			useMethod = method
		}
	}

	if useMethod == "" {
		return "", fmt.Errorf("There are available methods (%v), but none can be selected", methods)
	}

	err = ig.checkpointStep2(useMethod)
	if err != nil {
		return "", err
	}

	return useMethod, nil
}

var dataRegexp = regexp.MustCompile(`window._sharedData\s*=\s*(.+);`)

// step1: grab cookies and available login methods
func (ig *Instagram) checkpointStep1() ([]string, error) {
	body, cookies, err := ig.browserRequest("GET", ig.CheckpointURL, "", nil, "")
	if err != nil {
		return nil, err
	}

	res := dataRegexp.FindStringSubmatch(body)
	if len(res) != 2 {
		return nil, errors.New("Unknown format of challenge selection")
	}

	// there should be json actuality, but data is complex mostly useless, no need in full decode i think
	if !strings.Contains(res[1], "SelectVerificationMethodForm") {
		return nil, errors.New("Unknown format of challenge selection")
	}
	var methods []string
	if strings.Contains(res[1], `"email":`) {
		methods = append(methods, MethodEmail)
	}
	if strings.Contains(res[1], `"phone_number":`) {
		methods = append(methods, MethodSms)
	}

	if len(methods) == 0 {
		return nil, errors.New("Failed to determine any known chalenge methonds")
	}

	ig.CheckpointCookies = cookies

	return methods, nil
}

// step2: send code using given method
func (ig *Instagram) checkpointStep2(method string) error {
	challenges := map[string]struct {
		value    string
		formName string
	}{
		MethodSms: {
			"0", "VerifySMSCodeForm",
		},
		MethodEmail: {
			"1", "VerifyEmailCodeForm",
		},
	}

	ch, ok := challenges[method]
	if !ok {
		return errors.New("Incorrect method supplied")
	}

	values := map[string]string{
		"choice": ch.value,
	}

	body, cookies, err := ig.browserRequest("POST", ig.CheckpointURL, ig.CheckpointURL, ig.CheckpointCookies, encode(values))
	if err != nil {
		return err
	}

	if !strings.Contains(body, ch.formName) {
		return errors.New("Code input form not found")
	}

	ig.CheckpointCookies = cookies
	ig.Save()

	return nil
}

// CheckCodeF tries to submit instagram checkpont code
func (ig *Instagram) CheckpointStep3(code string) error {

	values := map[string]string{
		"security_code": code,
	}

	body, cookies, err := ig.browserRequest("POST", ig.CheckpointURL, ig.CheckpointURL, ig.CheckpointCookies, encode(values))
	if err != nil {
		return err
	}
	log.Debug(body)

	if !strings.Contains(body, "Your account has been verified.") || !strings.Contains(body, "Thanks!") {
		return errors.New("Bad code")
	}

	ig.CheckpointCookies = cookies
	return ig.checkpointStep4()
}

func (ig *Instagram) checkpointStep4() error {

	token, err := getToken(ig.CheckpointCookies)
	if err != nil {
		return err
	}

	// I wonder if Instagram devs made post parameters order matter INTENTIONALLY? If yes, they are fucken evil geniouses
	params := fmt.Sprintf("csrfmiddlewaretoken=%v&OK=OK", token)

	_, _, err = ig.browserRequest("POST", ig.CheckpointURL, ig.CheckpointURL, ig.CheckpointCookies, params)
	if err != nil {
		return err
	}

	ig.CheckpointCookies = nil
	ig.LoggedIn = true
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

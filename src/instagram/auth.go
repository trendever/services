package instagram

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Possible auth instagram errors
var (
	ErrorEmptyPassword = errors.New("Login or password is empty")
)

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

var uidRegexp = regexp.MustCompile(`https://i.instagram.com/integrity/checkpoint/checkpoint_logged_out_main/([0-9]+)/`)

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

// SendCode sends checkpoint code
func (ig *Instagram) SendCode(preferEmail bool) (string, error) {

	if ig.CheckpointURL == "" {
		return "", fmt.Errorf("Can not send code! Checkpoint URL is empty")
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

			/*_*/ useMethod = method
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

// step1: grab cookies and available login methods
func (ig *Instagram) checkpointStep1() ([]string, error) {

	body, cookies, err := ig.browserRequest("GET", ig.CheckpointURL, "", nil, "")
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

	body, cookies, err := ig.browserRequest("POST", ig.CheckpointURL, ig.CheckpointURL, ig.CheckpointCookies, encode(values))
	if err != nil {
		return err
	}

	if !strings.Contains(body, `<input id="id_response_code" inputmode="numeric" name="response_code"`) {
		return fmt.Errorf("Code input form not found")
	}

	ig.CheckpointCookies = cookies
	ig.Save()

	return nil
}

// CheckCodeF tries to submit instagram checkpont code
func (ig *Instagram) CheckpointStep3(code string) error {

	token, err := getToken(ig.CheckpointCookies)
	if err != nil {
		return err
	}

	// I wonder if Instagram devs made post parameters order matter INTENTIONALLY? If yes, they are fucken evil geniouses
	params := fmt.Sprintf("response_code=%v&csrfmiddlewaretoken=%v", code, token)

	body, cookies, err := ig.browserRequest("POST", ig.CheckpointURL, ig.CheckpointURL, ig.CheckpointCookies, params)
	if err != nil {
		return err
	}

	if !strings.Contains(body, "Your account has been verified.") || !strings.Contains(body, "Thanks!") {
		return fmt.Errorf("Bad code")
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

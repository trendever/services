package instagram

import (
	"common/log"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
)

// Possible auth instagram errors
var (
	ErrorEmptyPassword = errors.New("Login or password is empty")
)

type challengeStep struct {
	Message
	StepName  string      `json:"step_name"`
	UserID    uint64      `json:"user_id"`
	NonceCode string      `json:"nonce_code"`
	StepData  interface{} `json:"step_data"`
}

// Login to Instagram.
func (ig *Instagram) Login() error {
	if ig.Username == "" || ig.password == "" {
		return ErrorEmptyPassword
	}

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
		if loginResp.Message.IsCheckpoint() {
			ig.CheckpointURL = loginResp.Message.Challenge.APIPath
			uid, err := ig.getUidByCheckpointLink()
			if err != nil {
				return err
			}
			ig.UserID = uid
			ig.updateRankToken()
			return ErrorCheckpointRequired
		}
		return errors.New(loginResp.Message.ErrorType)
	}

	ig.UserID = loginResp.LoggedInUser.Pk
	ig.updateRankToken()
	ig.LoggedIn = true

	return nil
}

var uidRegexp = regexp.MustCompile(`/challenge/([0-9]+)/`)

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
		return "", errors.New("Can not send code! Checkpoint URL is empty")
	}

	choice, err := ig.selectChallenge(preferEmail)
	if err != nil {
		return "", err
	}

	method, err := ig.requestCode(choice)
	if err != nil {
		return "", err
	}

	return method, nil
}

// select challenge
func (ig *Instagram) selectChallenge(preferEmail bool) (choice string, err error) {
	// @TODO it's hard to tell how it will look for multiple methods
	var stepData struct {
		Choice string `json:"choice"`
		Email  string `json:"email"`

		// @CHECK this is just a guess. I do not have any accounts with phone
		Phone string `json:"phone"`

		FbAccessToken    string `json:"fb_access_token"`
		BigBlueToken     string `json:"big_blue_token"`
		GoogleOauthToken string `json:"google_oauth_token"`
	}

	reply := challengeStep{
		StepData: &stepData,
	}

	cookies, err := ig.loginRequest("GET", ig.CheckpointURL, "", &reply)
	if err != nil {
		return "", err
	}

	if reply.StepName != "select_verify_method" {
		return "", fmt.Errorf("unexpected challenge step '%v'", reply.StepName)
	}

	ig.Cookies = cookies

	return stepData.Choice, nil
}

func (ig *Instagram) requestCode(choice string) (method string, err error) {
	var stepData struct {
		ResendDelay  int    `json:"resend_delay"`
		ContactPoint string `json:"contact_point"`
		FormType     string `json:"form_type"`
	}
	reply := challengeStep{
		StepData: &stepData,
	}

	values := map[string]string{
		"choice": choice,
	}
	cookies, err := ig.loginRequest("POST", ig.CheckpointURL, encode(values), &reply)
	if err != nil {
		return "", err
	}

	ig.Cookies = cookies

	return stepData.FormType, nil
}

// tries to submit instagram checkpont code
func (ig *Instagram) SubmitCode(code string) error {
	var reply Message

	values := map[string]string{
		"security_code": code,
	}

	log.Debug("data: %v", values)

	err := ig.jsonRequest(ig.CheckpointURL, values, &reply)
	if err != nil {
		return err
	}

	if reply.Status != "ok" {
		return errors.New(reply.Message)
	}
	//@CHECK Well, according to reply we should go to instagram://checkpoint/dismiss here.
	// But accidental login worked for me and i do not have accounts with checkpoints anymore..

	ig.CheckpointURL = ""
	ig.Cookies = nil
	ig.LoggedIn = false
	return nil
}

func encode(params map[string]string) string {
	vals := url.Values{}
	for k, v := range params {
		vals.Add(k, v)
	}

	return vals.Encode()
}

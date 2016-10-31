package instagram

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Login to Instagram.
func (ig *Instagram) Login() error {

	fetch := fmt.Sprintf("/si/fetch_headers/?challenge_type=signup&guid=%v", generateUUID(false))

	resp, err := ig.requestMain("GET", fetch, nil, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// get csrftoken
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrftoken" {
			ig.token = cookie.Value
		}
	}

	if ig.userName == "" || ig.password == "" {
		return fmt.Errorf("Empty username or password")
	}

	// login
	login := &Login{
		DeviceId:          ig.deviceID,
		PhoneId:           ig.phoneID,
		Guid:              ig.uuid,
		UserName:          ig.userName,
		Password:          ig.password,
		Csrftoken:         ig.token,
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

	// get new csrftoken
	for _, cookie := range cookies {
		if cookie.Name == "csrftoken" {
			ig.token = cookie.Value
		}
	}

	ig.cookies = cookies

	if loginResp.Status == "fail" {
		if loginResp.Message.Message == "checkpoint_required" {
			ig.checkpointURL = loginResp.CheckpointURL
			return ErrorCheckpointRequired
		}
		return errors.New(loginResp.Message.Message)
	}

	ig.userNameID = loginResp.LoggedInUser.Pk
	ig.rankToken = fmt.Sprintf("%d_%v", ig.userNameID, ig.uuid)
	ig.isLoggedIn = true

	return nil
}

// SendCode sends checkpoint code
func (ig *Instagram) SendCode() {

}

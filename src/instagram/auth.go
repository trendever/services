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

	token, err := getToken(resp.Cookies())
	if err != nil {
		return err
	}
	ig.Token = token

	if ig.Username == "" || ig.Password == "" {
		return fmt.Errorf("Empty username or password")
	}

	// login
	login := &Login{
		DeviceId:          ig.DeviceID,
		PhoneId:           ig.PhoneID,
		Guid:              ig.UUID,
		UserName:          ig.Username,
		Password:          ig.Password,
		Csrftoken:         ig.Token,
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
			ig.Token = cookie.Value
		}
	}

	ig.Cookies = cookies

	if loginResp.Status == "fail" {
		if loginResp.Message.Message == "checkpoint_required" {
			ig.CheckpointURL = loginResp.CheckpointURL
			return ErrorCheckpointRequired
		}
		return errors.New(loginResp.Message.Message)
	}

	ig.UserNameID = loginResp.LoggedInUser.Pk
	ig.RankToken = fmt.Sprintf("%d_%v", ig.UserNameID, ig.UUID)
	ig.LoggedIn = true

	return nil
}

// SendCode sends checkpoint code
func (ig *Instagram) SendCode(preferEmail bool) (string, error) {

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

// CheckCode tries to submit instagram checkpont code
func (ig *Instagram) CheckCode(code string) error {

	return ig.checkpointSubmit(code)
}

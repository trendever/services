package instagram

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type Instagram struct {
	userName   string
	password   string
	token      string
	isLoggedIn bool
	uuid       string
	deviceId   string
	phoneId    string
	userNameId int64
	rankToken  string
	cookies    []*http.Cookie
}

func NewInstagram(userName, password string) (*Instagram, error) {
	i := &Instagram{
		userName:   userName,
		password:   password,
		token:      "",
		isLoggedIn: false,
		uuid:       generateUUID(true),
		phoneId:    generateUUID(true),
		deviceId:   generateDeviceID(userName),
		userNameId: 0,
		rankToken:  "",
		cookies:    nil,
	}

	err := i.Login()
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (this *Instagram) IsLoggedIn() bool {
	return this.isLoggedIn
}

func (this *Instagram) GetUserName() string {
	return this.userName
}

// Login to Instagram.
func (this *Instagram) Login() error {

	fetch := URL + "/si/fetch_headers/?challenge_type=signup&guid=" + generateUUID(false)

	resp, err := this.requestLogin("GET", fetch, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// get csrftoken
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrftoken" {
			this.token = cookie.Value
		}
	}

	// login
	login := &Login{
		DeviceId:          this.deviceId,
		PhoneId:           this.phoneId,
		Guid:              this.uuid,
		UserName:          this.userName,
		Password:          this.password,
		Csrftoken:         this.token,
		LoginAttemptCount: "0",
	}

	jsonData, err := json.Marshal(login)
	if err != nil {
		return err
	}

	signature := generateSignature(jsonData)

	resp, err = this.requestLogin("POST", URL+"/accounts/login/?", bytes.NewReader([]byte(signature)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// get new csrftoken
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrftoken" {
			this.token = cookie.Value
		}
	}
	this.cookies = resp.Cookies()

	var object *LoginResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&object)
	if err != nil {
		return err
	}

	if object.Status == "fail" {
		return errors.New(object.Message)
	}

	this.userNameId = object.LoggedInUser.Pk
	this.rankToken = strconv.FormatInt(this.userNameId, 10) + "_" + this.uuid
	this.isLoggedIn = true

	return nil
}

// Get media likers.
func (this *Instagram) GetMediaLikers(mediaId string) (*MediaLikers, error) {

	endpoint := URL + "/media/" + mediaId + "/likers/?"

	resp, err := this.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var object *MediaLikers
	err = json.Unmarshal(resp, &object)
	if err != nil {
		return nil, err
	}

	return object, nil
}

// Get media comments.
func (this *Instagram) GetMediaComment(mediaId string) (*MediaComment, error) {

	endpoint := URL + "/media/" + mediaId + "/comments/?"

	resp, err := this.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var object *MediaComment
	err = json.Unmarshal(resp, &object)
	if err != nil {
		return nil, err
	}

	return object, nil
}

// Get media by id.
func (this *Instagram) GetMedia(mediaId string) (*Medias, error) {

	endpoint := URL + "/media/" + mediaId + "/info/?"

	resp, err := this.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var object *Medias
	err = json.Unmarshal(resp, &object)
	if err != nil {
		return nil, err
	}

	return object, nil
}

// Get recent activity.
func (this *Instagram) GetRecentActivity() (*RecentActivity, error) {

	endpoint := URL + "/news/inbox/?"

	resp, err := this.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var object *RecentActivity
	err = json.Unmarshal(resp, &object)
	if err != nil {
		return nil, err
	}

	return object, nil
}

// Search users.
func (this *Instagram) SearchUsers(query string) (*SearchUsers, error) {

	endpoint := URL + "/users/search/?ig_sig_key_version=" + SigKeyVersion +
		"&is_typeahead=true&query=" + query + "&rank_token=" + this.rankToken

	resp, err := this.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var object *SearchUsers
	err = json.Unmarshal(resp, &object)
	if err != nil {
		return nil, err
	}

	return object, nil
}

// Get username info.
func (this *Instagram) GetUserNameInfo(userNameId int64) (*UserNameInfo, error) {

	endpoint := URL + "/users/" + strconv.FormatInt(userNameId, 10) + "/info/?"

	resp, err := this.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var object *UserNameInfo
	err = json.Unmarshal(resp, &object)
	if err != nil {
		return nil, err
	}

	return object, nil
}

// Get user tags.
func (this *Instagram) GetUserTags(userNameId int64) (*UserTags, error) {

	endpoint := URL + "/usertags/" + strconv.FormatInt(userNameId, 10) + "/feed/?rank_token=" +
		this.rankToken + "&ranked_content=false"

	resp, err := this.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var object *UserTags
	err = json.Unmarshal(resp, &object)
	if err != nil {
		return nil, err
	}

	return object, nil
}

// Search tags.
func (this *Instagram) SearchTags(query string) (*SearchTags, error) {

	endpoint := URL + "/tags/search/?is_typeahead=true&q=" + query + "&rank_token=" + this.rankToken

	resp, err := this.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var object *SearchTags
	err = json.Unmarshal(resp, &object)
	if err != nil {
		return nil, err
	}

	return object, nil
}

// Get tagged media.
func (this *Instagram) TagFeed(tag, maxId string) (*TagFeed, error) {

	endpoint := URL + "/feed/tag/" + tag + "/?rank_token=" + this.rankToken + "&ranked_content=false&max_id=" + maxId

	resp, err := this.request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var object *TagFeed
	err = json.Unmarshal(resp, &object)
	if err != nil {
		return nil, err
	}

	return object, nil
}

// Request for Login method. Needs to get the authorization cookies.
func (this *Instagram) requestLogin(method, endpoint string, body io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", UserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Accept-Language", "en-US")
	req.Header.Add("Cookie2", "$Version=1")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Main request for all other methods. Reading the authorization cookies.
func (this *Instagram) requestMain(method, endpoint string, body io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", UserAgent)
	for _, cookie := range this.cookies {
		req.AddCookie(cookie)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Request with five attempts re-login. Re-login if getting error 'login_required'.
func (this *Instagram) request(method, endpoint string, body io.Reader) ([]byte, error) {

	for attempt := 0; attempt < 5; attempt++ {

		resp, err := this.requestMain(method, endpoint, body)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		jsonBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var message *Message
		err = json.Unmarshal(jsonBody, &message)
		if err != nil {
			return nil, err
		}

		if message.Status == "fail" {
			if message.Message != "login_required" {
				return nil, errors.New(message.Message)
			}
			// relogin
			this.isLoggedIn = false
			err = this.Login()
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
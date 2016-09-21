package instagram

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Instagram defines client
type Instagram struct {
	userName   string
	password   string
	token      string
	isLoggedIn bool
	uuid       string
	deviceID   string
	phoneID    string
	userNameID int64
	rankToken  string
	cookies    []*http.Cookie
}

// NewInstagram initializes client for futher use
func NewInstagram(userName, password string) (*Instagram, error) {
	i := &Instagram{
		userName:   userName,
		password:   password,
		token:      "",
		isLoggedIn: false,
		uuid:       generateUUID(true),
		phoneID:    generateUUID(true),
		deviceID:   generateDeviceID(userName),
		userNameID: 0,
		rankToken:  "",
		cookies:    nil,
	}

	err := i.Login()
	if err != nil {
		return nil, err
	}

	return i, nil
}

// IsLoggedIn returns if last request does not have auth error
func (ig *Instagram) IsLoggedIn() bool {
	return ig.isLoggedIn
}

// GetUserName (will you guess what?) returns set-up username
func (ig *Instagram) GetUserName() string {
	return ig.userName
}

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
		return errors.New(loginResp.Message)
	}

	ig.userNameID = loginResp.LoggedInUser.Pk
	ig.rankToken = fmt.Sprintf("%d_%v", ig.userNameID, ig.uuid)
	ig.isLoggedIn = true

	return nil
}

// GetMediaLikers returns likers of given media
func (ig *Instagram) GetMediaLikers(mediaID string) (*MediaLikers, error) {

	endpoint := fmt.Sprintf("/media/%v/likers/?", mediaID)

	var object MediaLikers
	err := ig.request("GET", endpoint, &object)

	return &object, err
}

// GetMediaComment returns comment info for this media
func (ig *Instagram) GetMediaComment(mediaID string) (*MediaComment, error) {

	endpoint := fmt.Sprintf("/media/%v/comments/?", mediaID)

	var object MediaComment
	err := ig.request("GET", endpoint, &object)

	return &object, err
}

// GetMedia by id.
func (ig *Instagram) GetMedia(mediaID string) (*Medias, error) {

	endpoint := fmt.Sprintf("/media/%v/info/?", mediaID)

	var object Medias
	err := ig.request("GET", endpoint, &object)

	return &object, err
}

// GetRecentActivity returns main instagram feed
func (ig *Instagram) GetRecentActivity() (*RecentActivity, error) {

	endpoint := "/news/inbox/?"

	var object RecentActivity
	err := ig.request("GET", endpoint, &object)

	return &object, err
}

// SearchUsers find user by query
func (ig *Instagram) SearchUsers(query string) (*SearchUsers, error) {

	endpoint := fmt.Sprintf(
		"/users/search/?ig_sig_key_version=%v&is_typeahead=true&query=%v&rank_token=%v",
		SigKeyVersion, query, ig.rankToken,
	)

	var object SearchUsers
	err := ig.request("GET", endpoint, &object)

	return &object, err
}

// GetUserNameInfo for userNameID
func (ig *Instagram) GetUserNameInfo(userNameID int64) (*UserNameInfo, error) {

	endpoint := fmt.Sprintf("/users/%d/info/?", userNameID)

	var object UserNameInfo
	err := ig.request("GET", endpoint, &object)

	return &object, err
}

// GetUserTags for given userID
func (ig *Instagram) GetUserTags(userNameID int64) (*UserTags, error) {

	endpoint := fmt.Sprintf(
		"/usertags/%d/feed/?rank_token=%v&ranked_content=false",
		userNameID, ig.rankToken,
	)

	var object UserTags
	err := ig.request("GET", endpoint, &object)

	return &object, err
}

// SearchTags for some query
func (ig *Instagram) SearchTags(query string) (*SearchTags, error) {

	endpoint := fmt.Sprintf(
		"/tags/search/?is_typeahead=true&q=%v&rank_token=%v",
		query, ig.rankToken,
	)

	var object SearchTags
	err := ig.request("GET", endpoint, &object)

	return &object, err
}

// TagFeed returns tagged media.
func (ig *Instagram) TagFeed(tag, maxID string) (*TagFeed, error) {

	endpoint := fmt.Sprintf(
		"/feed/tag/%v/?rank_token=%v&ranked_content=false&max_id=%v",
		tag, ig.rankToken, maxID,
	)

	var object TagFeed
	err := ig.request("GET", endpoint, &object)

	return &object, err
}

// PendingInbox returns chats that needs to be either confirmed or declined
func (ig *Instagram) PendingInbox() (*PendingInboxResponse, error) {

	endpoint := "/direct_v2/pending_inbox/?"

	var object PendingInboxResponse
	err := ig.request("GET", endpoint, &object)

	return &object, err
}

// Inbox returns usual normal chats
func (ig *Instagram) Inbox() (*InboxResponse, error) {

	endpoint := "/direct_v2/inbox/?"

	var object InboxResponse
	err := ig.request("GET", endpoint, &object)

	return &object, err
}

// RankedRecipients returns @TODO wtf it returns?
func (ig *Instagram) RankedRecipients() (*RankedRecipientsResponse, error) {
	endpoint := "/direct_v2/ranked_recipients/?show_threads=true"

	var object RankedRecipientsResponse
	err := ig.request("GET", endpoint, &object)

	return &object, err

}

// DirectThread @TODO wtf returns
func (ig *Instagram) DirectThread(threadID string) (*DirectThreadResponse, error) {

	endpoint := fmt.Sprintf("/direct_v2/threads/%v/?", threadID)

	var object DirectThreadResponse
	err := ig.request("GET", endpoint, &object)

	return &object, err

}

// possible direct thread actions
const (
	ActionApprove = "approve"
	ActionDecline = "decline"
	ActionBlock   = "block"
)

// DirectThreadAction allows to accept or decline private thread
func (ig *Instagram) DirectThreadAction(threadID, action string) (*DirectThreadActionResponse, error) {

	endpoint := fmt.Sprintf("/direct_v2/threads/%v/%v/", threadID, action)

	var object DirectThreadActionResponse
	err := ig.request("POST", endpoint, &object)

	return &object, err
}

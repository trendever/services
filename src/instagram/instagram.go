package instagram

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Instagram defines client
type Instagram struct {
	Username          string
	Password          string
	Token             string
	LoggedIn          bool
	UUID              string
	DeviceID          string
	PhoneID           string
	UserNameID        int64
	RankToken         string
	CheckpointURL     string
	Cookies           []*http.Cookie
	CheckpointCookies []*http.Cookie
}

// NewInstagram initializes client for futher use
func NewInstagram(userName, password string) (*Instagram, error) {
	i := &Instagram{
		Username:   userName,
		Password:   password,
		Token:      "",
		LoggedIn:   false,
		UUID:       generateUUID(true),
		PhoneID:    generateUUID(true),
		DeviceID:   generateDeviceID(userName),
		UserNameID: 0,
		RankToken:  "",
		Cookies:    nil,
	}

	err := i.Login()
	return i, err
}

// IsLoggedIn returns if last request does not have auth error
func (ig *Instagram) IsLoggedIn() bool {
	return ig.LoggedIn
}

// GetUserName (will you guess what?) returns set-up username
func (ig *Instagram) GetUserName() string {
	return ig.Username
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
		SigKeyVersion, query, ig.RankToken,
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
		userNameID, ig.RankToken,
	)

	var object UserTags
	err := ig.request("GET", endpoint, &object)

	return &object, err
}

// SearchTags for some query
func (ig *Instagram) SearchTags(query string) (*SearchTags, error) {

	endpoint := fmt.Sprintf(
		"/tags/search/?is_typeahead=true&q=%v&rank_token=%v",
		query, ig.RankToken,
	)

	var object SearchTags
	err := ig.request("GET", endpoint, &object)

	return &object, err
}

// TagFeed returns tagged media.
func (ig *Instagram) TagFeed(tag, maxID string) (*TagFeed, error) {

	endpoint := fmt.Sprintf(
		"/feed/tag/%v/?rank_token=%v&ranked_content=false&max_id=%v",
		tag, ig.RankToken, maxID,
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
func (ig *Instagram) Inbox(cursor string) (*InboxResponse, error) {

	endpoint := "/direct_v2/inbox/?"
	if cursor != "" {
		endpoint += fmt.Sprintf("cursor=%v", url.QueryEscape(cursor))
	}

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
func (ig *Instagram) DirectThread(threadID, cursor string) (*DirectThreadResponse, error) {

	endpoint := fmt.Sprintf("/direct_v2/threads/%v/?", threadID)
	if cursor != "" {
		endpoint += fmt.Sprintf("cursor=%v", url.QueryEscape(cursor))
	}

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

// DirectThreadApproveAll allows to accept all the threads
func (ig *Instagram) DirectThreadApproveAll() (*DirectThreadApproveAllResponse, error) {

	endpoint := "/direct_v2/threads/approve_all/?"

	var object DirectThreadApproveAllResponse
	err := ig.request("POST", endpoint, &object)

	return &object, err
}

//BroadcastText sends text to given chat
func (ig *Instagram) BroadcastText(threadID, message string) (messageID string, _ error) {

	endpoint := "/direct_v2/threads/broadcast/text/"

	var object BroadcastTextResponse
	err := ig.postRequest(endpoint, map[string]string{
		"text":       message,
		"thread_ids": fmt.Sprintf("[%v]", threadID),
	}, &object)

	if err != nil {
		return "", err
	}

	if object.Message.Message != "" {
		return "", errors.New(object.Message.Message)
	}

	return object.Threads[0].NewestCursor, nil
}

// SendText sends text to given user
func (ig *Instagram) SendText(userID uint64, message string) (*SendTextResponse, error) {

	endpoint := "/direct_v2/threads/broadcast/text/"

	var object SendTextResponse
	err := ig.postRequest(endpoint, map[string]string{
		"text":            message,
		"recipient_users": fmt.Sprintf("[[%v]]", userID),
	}, &object)

	return &object, err

}

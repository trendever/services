package instagram

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type DialFunc func(network, addr string) (net.Conn, error)

// Instagram defines client
type Instagram struct {
	Username          string
	password          string
	LoggedIn          bool
	UserID            uint64 `json:"UserNameID"`
	RankToken         string
	CheckpointURL     string
	Cookies           []*http.Cookie
	CheckpointCookies []*http.Cookie
	Dial              DialFunc `json:"-"`

	// needed only for auth, but leave it there to make auth more stable
	UUID     string
	DeviceID string
	PhoneID  string
}

// NewInstagram initializes client for futher use
func NewInstagram(userName, password string, dial DialFunc) (*Instagram, error) {
	i := &Instagram{
		Username:  userName,
		password:  password,
		LoggedIn:  false,
		UUID:      generateUUID(true),
		PhoneID:   generateUUID(true),
		DeviceID:  generateDeviceID(userName),
		UserID:    0,
		RankToken: "",
		Cookies:   nil,
		Dial:      dial,
	}

	err := i.Login()
	return i, err
}

// IsLoggedIn returns if last request does not have auth error
func (ig *Instagram) IsLoggedIn() bool {
	return ig.LoggedIn
}

// SetPassword for this instance (use for restored accs)
func (ig *Instagram) SetPassword(pwd string) {
	ig.password = pwd
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

// CommentMedia returns comment info for this media
func (ig *Instagram) CommentMedia(mediaID, text string) (*Message, error) {

	// @TODO testing not to break login and checkpoints
	_, err := ig.SyncFeatures()
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/media/%v/comment/", mediaID)
	hashedstring := fmt.Sprintf("%x", md5.Sum([]byte(text)))

	var object Message
	err = ig.postRequest(endpoint, map[string]string{
		"idempotence_token": hashedstring,
		"src":               "profile",
		"comment_text":      text,
		"media_id":          mediaID,
	}, &object)

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

// GetUserFeed
func (ig *Instagram) GetUserFeed(userID uint64) (*UserFeed, error) {

	endpoint := fmt.Sprintf("/feed/user/%v/", userID)
	// @TODO: use max_id parameter

	var object UserFeed
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
func (ig *Instagram) GetUserNameInfo(userNameID uint64) (*UserNameInfo, error) {

	endpoint := fmt.Sprintf("/users/%d/info/?", userNameID)

	var object UserNameInfo
	err := ig.request("GET", endpoint, &object)

	return &object, err
}

// GetUserTags for given userID
func (ig *Instagram) GetUserTags(userNameID uint64) (*UserTags, error) {

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

// DirectThread
func (ig *Instagram) DirectThread(threadID, cursor string) (*DirectThreadResponse, error) {

	endpoint := fmt.Sprintf("/direct_v2/threads/%v/?", threadID)
	if cursor != "" {
		endpoint += fmt.Sprintf("cursor=%v", url.QueryEscape(cursor))
	}

	var object DirectThreadResponse
	err := ig.request("GET", endpoint, &object)

	return &object, err
}

// possible direct thread actions(there is some more actions actuality)
const (
	ActionApprove = "approve"
	ActionDecline = "decline"
	ActionBlock   = "block"
	ActionLeave   = "leave"
)

// DirectThreadAction allows to perform described above actions on thread
func (ig *Instagram) DirectThreadAction(threadID, action string) (*DirectThreadActionResponse, error) {

	endpoint := fmt.Sprintf("/direct_v2/threads/%v/%v/", threadID, action)

	var object DirectThreadActionResponse
	err := ig.request("POST", endpoint, &object)

	return &object, err
}

func (ig *Instagram) DirectUpdateTitle(threadID, title string) (*Message, error) {

	endpoint := fmt.Sprintf("/direct_v2/threads/%v/update_title/", threadID)

	var object Message
	err := ig.postRequest(endpoint, map[string]string{
		"title": title,
	}, &object)

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

	var object BroadcastResponse
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

	if len(object.Threads) == 0 {
		return "", errors.New("no threads info in responce")
	}

	return object.Threads[0].NewestCursor, nil
}

// SendText sends text to given user(-s)
func (ig *Instagram) SendText(message string, userIDs ...uint64) (threadID string, messageID string, err error) {
	endpoint := "/direct_v2/threads/broadcast/text/"

	strs := make([]string, len(userIDs), len(userIDs))
	for i, id := range userIDs {
		strs[i] = strconv.FormatUint(id, 10)
	}

	var object BroadcastResponse
	err = ig.postRequest(endpoint, map[string]string{
		"text":            message,
		"recipient_users": fmt.Sprintf("[[%v]]", strings.Join(strs, ",")),
	}, &object)

	if err != nil {
		return "", "", err
	}

	if object.Message.Message != "" {
		return "", "", errors.New(object.Message.Message)
	}

	if len(object.Threads) == 0 {
		return "", "", errors.New("no threads info in responce")
	}

	return object.Threads[0].ThreadID, object.Threads[0].NewestCursor, nil
}

func (ig *Instagram) SyncFeatures() (*Message, error) {
	endpoint := "/qe/sync/"

	token, err := getToken(ig.Cookies)
	if err != nil {
		return nil, err
	}

	var object Message
	err = ig.postRequest(endpoint, map[string]string{
		"_uuid":       ig.UUID,
		"_uid":        fmt.Sprintln(ig.UserID),
		"id":          fmt.Sprintln(ig.UserID),
		"_csrftoken":  token,
		"experiments": Experiments,
	}, &object)

	return &object, err

}

func (ig *Instagram) ShareMedia(threadID, mediaID string) (messageID string, _ error) {
	endpoint := "/direct_v2/threads/broadcast/media_share/?media_type=photo"

	var object BroadcastResponse
	err := ig.postRequest(endpoint, map[string]string{
		"media_id":   fmt.Sprintf("%v", mediaID),
		"thread_ids": fmt.Sprintf("[%v]", threadID),
	}, &object)

	if err != nil {
		return "", err
	}

	if object.Message.Message != "" {
		return "", errors.New(object.Message.Message)
	}

	if len(object.Threads) == 0 {
		return "", errors.New("no threads info in responce")
	}

	return object.Threads[0].NewestCursor, nil
}

// source should contain jpeg with aspect ration inside range 1.91:1 - 4:5, width limit is 1080px
func (ig *Instagram) SendPhoto(threadID string, source io.Reader) (messageID string, _ error) {
	endpoint := "/direct_v2/threads/broadcast/upload_photo/"

	params := map[string]string{
		// Format of response is different if this field is present. It is much shorter and probably falser.
		// @TODO test speed and update other methods
		"action":     "send_item",
		"thread_ids": fmt.Sprintf("[%v]", threadID),
	}

	var buf bytes.Buffer
	multi := multipart.NewWriter(&buf)
	for name, value := range params {
		multi.WriteField(name, value)
	}
	header := make(textproto.MIMEHeader)
	header.Set(
		"Content-Disposition",
		fmt.Sprintf(`form-data; name="photo"; filename="direct_temp_photo_%v.jpeg"`, time.Now().UnixNano()/1000),
	)
	header.Set("Content-Type", "application/octet-stream")
	header.Set("Content-Transfer-Encoding", "binary")
	writer, err := multi.CreatePart(header)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(writer, source)
	if err != nil {
		return "", err
	}
	multi.Close()

	var object DirectPhotoResponse
	err = ig.postContentRequest(endpoint, &PostContent{
		Type: multi.FormDataContentType(), Reader: &buf,
	}, &object)
	if err != nil {
		return "", err
	}

	return object.Payload.ItemID, nil
}

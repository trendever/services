package instagram

// PendingInboxResponse contains chats that are needed to be either confirmed or declined
type PendingInboxResponse struct {
	Message
	SeqID                int `json:"seq_id"`
	PendingRequestsTotal int `json:"pending_requests_total"`
	Inbox                struct {
		UnseenCount   int      `json:"unseen_count"`
		HasOlder      bool     `json:"has_older"`
		UnseenCountTs int64    `json:"unseen_count_ts"`
		Threads       []Thread `json:"threads"`
	} `json:"inbox"`
}

// RankedRecipientsResponse returns list of open chats
type RankedRecipientsResponse struct {
	Message
	Expires          int `json:"expires"`
	RankedRecipients []struct {
		Thread struct {
			Named bool `json:"named"`
			Users []struct {
				Username                   string `json:"username"`
				HasAnonymousProfilePicture bool   `json:"has_anonymous_profile_picture"`
				ProfilePicURL              string `json:"profile_pic_url"`
				ProfilePicID               string `json:"profile_pic_id"`
				FullName                   string `json:"full_name"`
				Pk                         int64  `json:"pk"`
				IsPrivate                  bool   `json:"is_private"`
			} `json:"users"`
			ThreadType  string `json:"thread_type"`
			ThreadID    string `json:"thread_id"`
			ThreadTitle string `json:"thread_title"`
			Pending     bool   `json:"pending"`
		} `json:"thread"`
	} `json:"ranked_recipients"`
}

// InboxResponse returns full inbox
type InboxResponse struct {
	Message
	PendingRequestsTotal int           `json:"pending_requests_total"`
	SeqID                int           `json:"seq_id"`
	PendingRequestsUsers []interface{} `json:"pending_requests_users"`
	Inbox                struct {
		UnseenCount   int      `json:"unseen_count"`
		HasOlder      bool     `json:"has_older"`
		OldestCursor  string   `json:"oldest_cursor"`
		UnseenCountTs int64    `json:"unseen_count_ts"`
		Threads       []Thread `json:"threads"`
	} `json:"inbox"`
	Subscription struct {
		Topic    string `json:"topic"`
		URL      string `json:"url"`
		Auth     string `json:"auth"`
		Sequence string `json:"sequence"`
	} `json:"subscription"`
}

// DirectThreadResponse contains the whole thread
type DirectThreadResponse struct {
	Message
	Thread Thread `json:"thread"`
}

// Thread is direct thread
type Thread struct {
	Named bool `json:"named"`
	Users []struct {
		Username                   string `json:"username"`
		HasAnonymousProfilePicture bool   `json:"has_anonymous_profile_picture"`
		FriendshipStatus           struct {
			Following       bool `json:"following"`
			IncomingRequest bool `json:"incoming_request"`
			OutgoingRequest bool `json:"outgoing_request"`
			Blocking        bool `json:"blocking"`
			IsPrivate       bool `json:"is_private"`
		} `json:"friendship_status"`
		ProfilePicURL string `json:"profile_pic_url"`
		FullName      string `json:"full_name"`
		Pk            uint64 `json:"pk"`
		IsVerified    bool   `json:"is_verified"`
		IsPrivate     bool   `json:"is_private"`
	} `json:"users"`
	HasNewer       bool           `json:"has_newer"`
	ThreadID       string         `json:"thread_id"`
	ImageVersions2 ImageVersions2 `json:"image_versions2"`
	LastActivityAt int64          `json:"last_activity_at"`
	NewestCursor   string         `json:"newest_cursor"`
	Canonical      bool           `json:"canonical"`
	HasOlder       bool           `json:"has_older"`
	OldestCursor   string         `json:"oldest_cursor"`
	LeftUsers      []interface{}  `json:"left_users"`
	Muted          bool           `json:"muted"`
	Items          ThreadItems    `json:"items"`
	// no longer exists it seams
	ThreadType  string `json:"thread_type"`
	ThreadTitle string `json:"thread_title"`
	LastSeenAt  map[string]struct {
		ItemID    string `json:"item_id"`
		Timestamp string `json:"timestamp"`
	} `json:"last_seen_at"`
	Inviter struct {
		Username                   string `json:"username"`
		HasAnonymousProfilePicture bool   `json:"has_anonymous_profile_picture"`
		ProfilePicURL              string `json:"profile_pic_url"`
		FullName                   string `json:"full_name"`
		Pk                         int64  `json:"pk"`
		IsVerified                 bool   `json:"is_verified"`
		IsPrivate                  bool   `json:"is_private"`
	} `json:"inviter"`
	Pending bool `json:"pending"`
}

// MediaShare contains shared stuff
type MediaShare struct {
	TakenAt         int            `json:"taken_at"`
	Pk              int64          `json:"pk"`
	ID              string         `json:"id"`
	DeviceTimestamp int64          `json:"device_timestamp"`
	MediaType       int            `json:"media_type"`
	Code            string         `json:"code"`
	ClientCacheKey  string         `json:"client_cache_key"`
	FilterType      int            `json:"filter_type"`
	ImageVersions2  ImageVersions2 `json:"image_versions2"`
	OriginalWidth   int            `json:"original_width"`
	OriginalHeight  int            `json:"original_height"`
	User            struct {
		Username                   string `json:"username"`
		HasAnonymousProfilePicture bool   `json:"has_anonymous_profile_picture"`
		IsUnpublished              bool   `json:"is_unpublished"`
		ProfilePicURL              string `json:"profile_pic_url"`
		FullName                   string `json:"full_name"`
		Pk                         uint64 `json:"pk"`
		IsPrivate                  bool   `json:"is_private"`
	} `json:"user"`
	OrganicTrackingToken         string        `json:"organic_tracking_token"`
	LikeCount                    int           `json:"like_count"`
	Likers                       []interface{} `json:"likers"`
	HasLiked                     bool          `json:"has_liked"`
	HasMoreComments              bool          `json:"has_more_comments"`
	MaxNumVisiblePreviewComments int           `json:"max_num_visible_preview_comments"`
	Comments                     []struct {
		Status       string `json:"status"`
		UserID       int64  `json:"user_id"`
		CreatedAtUtc int    `json:"created_at_utc"`
		CreatedAt    int    `json:"created_at"`
		BitFlags     int    `json:"bit_flags"`
		User         struct {
			Username      string `json:"username"`
			ProfilePicURL string `json:"profile_pic_url"`
			ProfilePicID  string `json:"profile_pic_id"`
			FullName      string `json:"full_name"`
			Pk            int64  `json:"pk"`
			IsVerified    bool   `json:"is_verified"`
			IsPrivate     bool   `json:"is_private"`
		} `json:"user"`
		ContentType    string `json:"content_type"`
		Text           string `json:"text"`
		MediaID        int64  `json:"media_id"`
		Pk             int64  `json:"pk"`
		HasTranslation bool   `json:"has_translation"`
		Type           int    `json:"type"`
	} `json:"comments"`
	CommentCount int `json:"comment_count"`
	Caption      struct {
		Status       string `json:"status"`
		UserID       int64  `json:"user_id"`
		CreatedAtUtc int    `json:"created_at_utc"`
		CreatedAt    int    `json:"created_at"`
		BitFlags     int    `json:"bit_flags"`
		User         struct {
			Username                   string `json:"username"`
			HasAnonymousProfilePicture bool   `json:"has_anonymous_profile_picture"`
			IsUnpublished              bool   `json:"is_unpublished"`
			ProfilePicURL              string `json:"profile_pic_url"`
			FullName                   string `json:"full_name"`
			Pk                         int64  `json:"pk"`
			IsPrivate                  bool   `json:"is_private"`
		} `json:"user"`
		ContentType string `json:"content_type"`
		Text        string `json:"text"`
		MediaID     int64  `json:"media_id"`
		Pk          int64  `json:"pk"`
		Type        int    `json:"type"`
	} `json:"caption"`
	CaptionIsEdited bool `json:"caption_is_edited"`
	PhotoOfYou      bool `json:"photo_of_you"`
}

// DirectThreadActionResponse if the request is ok
type DirectThreadActionResponse struct {
	Message
}

// DirectThreadApproveAllResponse if the request is ok
type DirectThreadApproveAllResponse struct {
	Message
}

type BroadcastResponse struct {
	Message
	Threads []Thread `json:"threads"`
}

type DirectPhotoResponse struct {
	Message
	Action  string `json:"action"`
	Payload struct {
		Timestamp     string `json:"timestamp"`
		ThreadID      string `json:"thread_id"`
		ItemID        string `json:"item_id"`
		ClientContext string `json:"client_context"`
		Message       string `json:"message"`
	} `json:"payload"`
}

type MediaType int

const (
	MediaType_Image    = 1
	MediaType_Video    = 2
	MediaType_Carousel = 8
)

type DirectMedia struct {
	MediaType      MediaType      `json:"media_type"`
	OriginalWidth  uint           `json:"original_width"`
	OriginalHeight uint           `json:"original_height"`
	ImageVersions2 ImageVersions2 `json:"image_versions2"`
	VideoVersions  []VideoVersion `json:"video_versions"`
	ExpiringAt     uint64         `json:"expiring_at"`
}

type Link struct {
	Text    string `json:"text"`
	Context struct {
		URL      string `json:"link_url"`
		Title    string `json:"link_title"`
		Summary  string `json:"link_summary"`
		ImageURL string `json:"link_image_url"`
	} `json:"link_context"`
}

type Location struct {
	ID               uint64  `json:"pk"`
	Name             string  `json:"name"`
	ShortName        string  `json:"short_name"`
	Address          string  `json:"address"`
	City             string  `json:"city"`
	Longitude        float32 `json:"lng"`
	Latitude         float32 `json:"lat"`
	ExternalSource   string  `json:"external_source"`
	FacebookPlacesID uint64  `json:"facebook_places_id"`
}

// ThreadItems contains messages from the chat
type ThreadItems []ThreadItem

// ThreadItem contains one message from the chat
type ThreadItem struct {
	UserID        uint64      `json:"user_id"`
	Text          string      `json:"text,omitempty"`
	ItemType      string      `json:"item_type"`
	Timestamp     int64       `json:"timestamp"`
	ItemID        string      `json:"item_id"`
	ClientContext string      `json:"client_context"`
	Media         DirectMedia `json:"media"`
	MediaShare    *MediaShare `json:"media_share,omitempty"`
	Link          Link        `json:"link"`
	Like          string      `json:"like"`
	Profile       *User       `json:"profile"`
	Location      *Location   `json:"location"`
	HashTag       struct {
		Name       string `json:"name"`
		MediaCount uint64 `json:"media_count"`
		// there is also some medias right in item, but there is no need to decode them
	} `json:"hashtag"`
	ReelShare *struct {
		Type  string      `json:"type"`
		Text  string      `json:"text"`
		Media DirectMedia `json:"media"`
	} `json:"reel_share"`
}

// Sorting stuff
func (a ThreadItems) Len() int           { return len(a) }
func (a ThreadItems) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ThreadItems) Less(i, j int) bool { return a[i].ItemID > a[j].ItemID } // messages with greater IDs first

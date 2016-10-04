package instagram

// PendingInboxResponse contains chats that are needed to be either confirmed or declined
type PendingInboxResponse struct {
	Message
	SeqID                int `json:"seq_id"`
	PendingRequestsTotal int `json:"pending_requests_total"`
	Inbox                struct {
		UnseenCount   int   `json:"unseen_count"`
		HasOlder      bool  `json:"has_older"`
		UnseenCountTs int64 `json:"unseen_count_ts"`
		Threads       []struct {
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
				ProfilePicID  string `json:"profile_pic_id"`
				FullName      string `json:"full_name"`
				Pk            int64  `json:"pk"`
				IsPrivate     bool   `json:"is_private"`
			} `json:"users"`
			HasNewer       bool          `json:"has_newer"`
			ThreadID       string        `json:"thread_id"`
			LastActivityAt int64         `json:"last_activity_at"`
			NewestCursor   string        `json:"newest_cursor"`
			IsSpuham       bool          `json:"is_spam"`
			HasOlder       bool          `json:"has_older"`
			OldestCursor   string        `json:"oldest_cursor"`
			LeftUsers      []interface{} `json:"left_users"`
			Muted          bool          `json:"muted"`
			Items          []struct {
				ItemID    string `json:"item_id"`
				ItemType  string `json:"item_type"`
				Text      string `json:"text"`
				UserID    int64  `json:"user_id"`
				Timestamp int64  `json:"timestamp"`
			} `json:"items"`
			ThreadType  string `json:"thread_type"`
			ThreadTitle string `json:"thread_title"`
			Canonical   bool   `json:"canonical"`
			Inviter     struct {
				Username                   string `json:"username"`
				HasAnonymousProfilePicture bool   `json:"has_anonymous_profile_picture"`
				ProfilePicURL              string `json:"profile_pic_url"`
				ProfilePicID               string `json:"profile_pic_id"`
				FullName                   string `json:"full_name"`
				Pk                         int64  `json:"pk"`
				IsPrivate                  bool   `json:"is_private"`
			} `json:"inviter"`
			Pending bool `json:"pending"`
		} `json:"threads"`
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
		UnseenCount   int   `json:"unseen_count"`
		HasOlder      bool  `json:"has_older"`
		UnseenCountTs int64 `json:"unseen_count_ts"`
		Threads       []struct {
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
				Pk            int64  `json:"pk"`
				IsPrivate     bool   `json:"is_private"`
			} `json:"users"`
			HasNewer       bool   `json:"has_newer"`
			ThreadID       string `json:"thread_id"`
			ImageVersions2 struct {
				Candidates []struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"candidates"`
			} `json:"image_versions2,omitempty"`
			LastActivityAt int64         `json:"last_activity_at"`
			NewestCursor   string        `json:"newest_cursor"`
			IsSpam         bool          `json:"is_spam"`
			HasOlder       bool          `json:"has_older"`
			OldestCursor   string        `json:"oldest_cursor"`
			LeftUsers      []interface{} `json:"left_users"`
			Muted          bool          `json:"muted"`
			Items          ThreadItems   `json:"items"`
			ThreadType     string        `json:"thread_type"`
			ThreadTitle    string        `json:"thread_title"`
			LastSeenAt     map[string]struct {
				ItemID    string `json:"item_id"`
				Timestamp string `json:"timestamp"`
			} `json:"last_seen_at,omitempty"`
			Canonical bool `json:"canonical"`
			Inviter   struct {
				Username                   string `json:"username"`
				HasAnonymousProfilePicture bool   `json:"has_anonymous_profile_picture"`
				ProfilePicURL              string `json:"profile_pic_url"`
				FullName                   string `json:"full_name"`
				Pk                         int64  `json:"pk"`
				IsPrivate                  bool   `json:"is_private"`
			} `json:"inviter"`
			Pending bool `json:"pending"`
		} `json:"threads"`
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
	Thread struct {
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
			Pk            int64  `json:"pk"`
			IsVerified    bool   `json:"is_verified"`
			IsPrivate     bool   `json:"is_private"`
		} `json:"users"`
		HasNewer       bool   `json:"has_newer"`
		ThreadID       string `json:"thread_id"`
		ImageVersions2 struct {
			Candidates []struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"candidates"`
		} `json:"image_versions2"`
		LastActivityAt int64         `json:"last_activity_at"`
		NewestCursor   string        `json:"newest_cursor"`
		Canonical      bool          `json:"canonical"`
		HasOlder       bool          `json:"has_older"`
		OldestCursor   string        `json:"oldest_cursor"`
		LeftUsers      []interface{} `json:"left_users"`
		Muted          bool          `json:"muted"`
		Items          ThreadItems   `json:"items"`
		ThreadType     string        `json:"thread_type"`
		ThreadTitle    string        `json:"thread_title"`
		LastSeenAt     map[string]struct {
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
	} `json:"thread"`
}

// MediaShare contains shared stuff
type MediaShare struct {
	TakenAt         int    `json:"taken_at"`
	Pk              int64  `json:"pk"`
	ID              string `json:"id"`
	DeviceTimestamp int64  `json:"device_timestamp"`
	MediaType       int    `json:"media_type"`
	Code            string `json:"code"`
	ClientCacheKey  string `json:"client_cache_key"`
	FilterType      int    `json:"filter_type"`
	ImageVersions2  struct {
		Candidates []struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"candidates"`
	} `json:"image_versions2"`
	OriginalWidth  int `json:"original_width"`
	OriginalHeight int `json:"original_height"`
	User           struct {
		Username                   string `json:"username"`
		HasAnonymousProfilePicture bool   `json:"has_anonymous_profile_picture"`
		IsUnpublished              bool   `json:"is_unpublished"`
		ProfilePicURL              string `json:"profile_pic_url"`
		FullName                   string `json:"full_name"`
		Pk                         int64  `json:"pk"`
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

// ThreadItems contains messages from chat
type ThreadItems []struct {
	UserID        int64       `json:"user_id"`
	Text          string      `json:"text,omitempty"`
	ItemType      string      `json:"item_type"`
	Timestamp     int64       `json:"timestamp"`
	ItemID        string      `json:"item_id"`
	ClientContext string      `json:"client_context"`
	MediaShare    *MediaShare `json:"media_share,omitempty"`
}

// Sorting stuff
func (a ThreadItems) Len() int           { return len(a) }
func (a ThreadItems) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ThreadItems) Less(i, j int) bool { return a[i].ItemID > a[j].ItemID } // messages with greater IDs first

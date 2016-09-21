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
			IsSpam         bool          `json:"is_spam"`
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

// DirectThreadResponse contains the whole thread
type DirectThreadResponse struct {
	Message
}

// DirectThreadActionResponse if the request is ok
type DirectThreadActionResponse struct {
	Message
}

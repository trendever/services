package instagram

type MediaComment struct {
	Comments []struct {
		Status    string  `json:"status"`
		MediaID   int64   `json:"media_id"`
		Text      string  `json:"text"`
		CreatedAt float64 `json:"created_at"`
		User      struct {
			Username                   string `json:"username"`
			HasAnonymousProfilePicture bool   `json:"has_anonymous_profile_picture"`
			ProfilePicURL              string `json:"profile_pic_url"`
			FullName                   string `json:"full_name"`
			Pk                         int64  `json:"pk"`
			IsPrivate                  bool   `json:"is_private"`
		} `json:"user"`
		ContentType  string `json:"content_type"`
		CreatedAtUtc int    `json:"created_at_utc"`
		Pk           int64  `json:"pk"`
		Type         int    `json:"type"`
	} `json:"comments"`
	Caption struct {
		Status       string `json:"status"`
		UserID       int64  `json:"user_id"`
		CreatedAtUtc int    `json:"created_at_utc"`
		CreatedAt    int    `json:"created_at"`
		BitFlags     int    `json:"bit_flags"`
		User         struct {
			Username      string `json:"username"`
			Pk            int64  `json:"pk"`
			ProfilePicURL string `json:"profile_pic_url"`
			IsPrivate     bool   `json:"is_private"`
			FullName      string `json:"full_name"`
		} `json:"user"`
		ContentType string `json:"content_type"`
		Text        string `json:"text"`
		Pk          int64  `json:"pk"`
		Type        int    `json:"type"`
	} `json:"caption"`
	CommentCount int `json:"comment_count"`
}

type Medias struct {
	Status          string      `json:"status"`
	CaptionIsEdited bool        `json:"caption_is_edited"`
	HasMoreComments bool        `json:"has_more_comments"`
	Items           []MediaInfo `json:"items"`
	Message         string      `json:"message"` // from Error
}

type MediaInfo struct {
	Caption struct {
		BitFlags     int    `json:"bit_flags"`
		ContentType  string `json:"content_type"`
		CreatedAt    int64  `json:"created_at"`
		CreatedAtUtc int64  `json:"created_at_utc"`
		MediaID      int64  `json:"media_id"`
		Pk           int    `json:"pk"`
		Status       string `json:"status"`
		Text         string `json:"text"`
		Type         int    `json:"type"`
		User         struct {
			FullName      string `json:"full_name"`
			IsPrivate     bool   `json:"is_private"`
			Pk            int    `json:"pk"`
			ProfilePicURL string `json:"profile_pic_url"`
			Username      string `json:"username"`
		} `json:"user"`
		UserID int64 `json:"user_id"`
	} `json:"caption"`
	CaptionIsEdited bool   `json:"caption_is_edited"`
	ClientCacheKey  string `json:"client_cache_key"`
	Code            string `json:"code"`
	CommentCount    int    `json:"comment_count"`
	Comments        []struct {
		BitFlags     int    `json:"bit_flags"`
		ContentType  string `json:"content_type"`
		CreatedAt    int    `json:"created_at"`
		CreatedAtUtc int    `json:"created_at_utc"`
		MediaID      int    `json:"media_id"`
		Pk           int    `json:"pk"`
		Status       string `json:"status"`
		Text         string `json:"text"`
		Type         int    `json:"type"`
		User         struct {
			FullName      string `json:"full_name"`
			IsPrivate     bool   `json:"is_private"`
			Pk            int    `json:"pk"`
			ProfilePicURL string `json:"profile_pic_url"`
			Username      string `json:"username"`
		} `json:"user"`
		UserID int `json:"user_id"`
	} `json:"comments"`
	FilterType      int    `json:"filter_type"`
	HasLiked        bool   `json:"has_liked"`
	HasMoreComments bool   `json:"has_more_comments"`
	ID              string `json:"id"`
	ImageVersions2  struct {
		Candidates []struct {
			Height int    `json:"height"`
			URL    string `json:"url"`
			Width  int    `json:"width"`
		} `json:"candidates"`
	} `json:"image_versions2"`
	LikeCount                    int           `json:"like_count"`
	Likers                       []interface{} `json:"likers"`
	MaxNumVisiblePreviewComments int           `json:"max_num_visible_preview_comments"`
	MediaType                    int           `json:"media_type"`
	NextMaxID                    int           `json:"next_max_id"`
	OrganicTrackingToken         string        `json:"organic_tracking_token"`
	OriginalHeight               int           `json:"original_height"`
	OriginalWidth                int           `json:"original_width"`
	PhotoOfYou                   bool          `json:"photo_of_you"`
	Pk                           int           `json:"pk"`
	TakenAt                      float64       `json:"taken_at"`
	User                         struct {
		FullName                   string `json:"full_name"`
		HasAnonymousProfilePicture bool   `json:"has_anonymous_profile_picture"`
		IsPrivate                  bool   `json:"is_private"`
		IsUnpublished              bool   `json:"is_unpublished"`
		Pk                         int    `json:"pk"`
		ProfilePicURL              string `json:"profile_pic_url"`
		Username                   string `json:"username"`
	} `json:"user"`
}

// Get media likers.
type MediaLikers struct {
	Status    string `json:"status"`
	UserCount int    `json:"user_count"`
	Users     []struct {
		Username      string `json:"username"`
		Pk            int64  `json:"pk"`
		ProfilePicURL string `json:"profile_pic_url"`
		IsPrivate     bool   `json:"is_private"`
		FullName      string `json:"full_name"`
	} `json:"users"`
	Message string `json:"message"` // from Error
}

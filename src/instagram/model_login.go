package instagram

// Login to Instagram.
type Login struct {
	DeviceId          string `json:"device_id"`
	PhoneId           string `json:"phone_id"`
	Guid              string `json:"guid"`
	UserName          string `json:"username"`
	Password          string `json:"password"`
	Csrftoken         string `json:"_csrftoken"`
	LoginAttemptCount string `json:"login_attempt_count"`
}

// Login to Instagram.
type LoginResponse struct {
	Message
	LoggedInUser struct {
		Username                   string `json:"username"`
		HasAnonymousProfilePicture bool   `json:"has_anonymous_profile_picture"`
		ProfilePicURL              string `json:"profile_pic_url"`
		FullName                   string `json:"full_name"`
		Pk                         uint64 `json:"pk"`
		IsPrivate                  bool   `json:"is_private"`
	} `json:"logged_in_user"`
}

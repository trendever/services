package instagram

import (
	"encoding/json"
	"utils/log"
)

// Save encodes connection to saveable string
func (ig *Instagram) Save() (string, error) {

	bytes, err := json.Marshal(&ig)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Restore previously saved
func Restore(cookieJar, password string, tryPing bool) (*Instagram, error) {

	var res Instagram
	err := json.Unmarshal([]byte(cookieJar), &res)
	if err != nil {
		return nil, err
	}

	if res.UserID <= 0 {
		log.Warn("Bad cookie: zero instagram ID (%v)", res.Username)
	}

	res.password = password

	// test request
	if tryPing {
		_, err = res.GetRecentActivity()
		if err != nil {
			return &res, err // we still need to give-away instagram
		}
	}

	return &res, nil
}

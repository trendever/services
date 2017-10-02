package instagram

import (
	"common/log"
	"encoding/json"
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
func Restore(cookieJar, password string, tryPing, responseLogging bool) (*Instagram, error) {

	var res Instagram
	err := json.Unmarshal([]byte(cookieJar), &res)
	if err != nil {
		return nil, err
	}

	if res.UserID <= 0 {
		log.Warn("Bad cookie: zero instagram ID (%v)", res.Username)
	}

	res.password = password
	res.transport, err = transportFromURL(res.Proxy)
	if err != nil {
		return nil, err
	}
	res.ResponseLogging = responseLogging

	// test request
	if tryPing {
		_, err = res.GetRecentActivity()
		if err != nil {
			return &res, err // we still need to give-away instagram
		}
	}

	return &res, nil
}

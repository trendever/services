package instagram

import (
	"encoding/json"
)

// Save encodes connection to saveable string
func (ig *Instagram) Save() (string, error) {

	var copy = *ig
	copy.password = ""

	bytes, err := json.Marshal(&copy)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Restore previously saved
func Restore(cookieJar, password string) (*Instagram, error) {

	var res Instagram
	err := json.Unmarshal([]byte(cookieJar), &res)
	if err != nil {
		return nil, err
	}

	res.password = password

	// test request
	_, err = res.GetRecentActivity()
	if err != nil {
		return nil, err
	}

	// clear checkpoint stuff
	res.CheckpointURL = ""
	res.CheckpointCookies = nil
	return &res, nil
}

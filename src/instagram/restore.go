package instagram

import (
	"encoding/json"
)

// Save encodes connection to saveable string
func (ig *Instagram) Save() (string, error) {

	var copy = *ig
	copy.Password = ""

	bytes, err := json.Marshal(&copy)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Restore previously saved
func Restore(cookieJar string) (*Instagram, error) {

	var res Instagram
	err := json.Unmarshal([]byte(cookieJar), &res)
	if err != nil {
		return nil, err
	}

	// @TODO: test ping

	return &res, nil
}

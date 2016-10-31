package instagram

import (
	"encoding/json"
)

// Save encodes connection to saveable string
func (ig *Instagram) Save() (string, error) {

	enc := map[string]interface{}{
		"userName":      ig.userName,
		"token":         ig.token,
		"uuid":          ig.uuid,
		"phoneID":       ig.phoneID,
		"deviceID":      ig.deviceID,
		"cookies":       ig.cookies,
		"userNameID":    ig.userNameID,
		"rankToken":     ig.rankToken,
		"checkpointURL": ig.checkpointURL,
	}

	bytes, err := json.Marshal(enc)
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

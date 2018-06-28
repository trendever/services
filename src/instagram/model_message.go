package instagram

// Message is what every instagram response based on
type Message struct {
	Status    string `json:"status"`
	Message   string `json:"message"` // from Error
	ErrorType string `json:"error_type"`
	Challenge struct {
		URL     string `json:"url"`
		APIPath string `json:"api_path"`
		Lock    bool   `json:"lock"`
		Logout  bool   `json:"logout"`
	} `json:"challenge"`
}

func (m Message) IsCheckpoint() bool {
	return m.ErrorType == "checkpoint_challenge_required"
}

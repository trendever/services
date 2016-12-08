package instagram

// Message is what every instagram response based on
type Message struct {
	Status        string `json:"status"`
	Message       string `json:"message"` // from Error
	CheckpointURL string `json:"checkpoint_url"`
}

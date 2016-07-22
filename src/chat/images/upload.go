package images

import (
	"bytes"
	"chat/config"
	"encoding/json"
	"errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
)

type ImageResp struct {
	Data    *Image `json:"data"`
	Status  int    `json:"status"`
	Success bool   `json:"success"`
}

type Image struct {
	Link    string                 `json:"link"`
	Mime    string                 `json:"mime"`
	Name    string                 `json:"name"`
	Hash    string                 `json:"hash"`
	Size    int64                  `json:"size"`
	Width   int                    `json:"width"`
	Height  int                    `json:"height"`
	OCRText string                 `json:"ocrtext"`
	Thumbs  map[string]interface{} `json:"thumbs"`
	UserID  string                 `json:"user_id"`
}

//UploadBase64 uploads base64 content
func UploadBase64(content string) (*Image, error) {
	u, err := url.Parse(config.Get().UploadService)
	if err != nil {
		return nil, err
	}

	u.Path = "/base64"

	reqBody := &bytes.Buffer{}

	writer := multipart.NewWriter(reqBody)
	if err := writer.WriteField("image", content); err != nil {
		return nil, err
	}
	thumbs := map[string]interface{}{
		"big": map[string]interface{}{
			"width":  1080,
			"height": 1080,
			"shape":  "thumb",
		},
		"small": map[string]interface{}{
			"width":  480,
			"height": 480,
			"shape":  "thumb",
		},
		"small_crop": map[string]interface{}{
			"width":  480,
			"height": 480,
			"shape":  "square",
		},
	}
	thumbsJSON, _ := json.Marshal(thumbs)

	if err := writer.WriteField("thumbs", string(thumbsJSON)); err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", u.String(), reqBody)
	req.Header.Add("Content-Type", "multipart/form-data; boundary="+writer.Boundary())

	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	m := &ImageResp{}

	err = json.Unmarshal(body, &m)
	if err != nil {
		return nil, err
	}
	if !m.Success {
		return nil, errors.New(string(body))
	}
	return m.Data, nil

}

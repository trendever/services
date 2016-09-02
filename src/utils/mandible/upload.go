package mandible

import (
	"encoding/json"
	"fmt"

	"github.com/derlaft/sling"
)

type thumbnailRequest map[string]*Thumbnail

// Thumbnail should contain info about wanted thumbnail
type Thumbnail struct {
	Name   string `json:"-"`
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
	Shape  string `json:"shape"`
}

type imageResp struct {
	Data    *image `json:"data"`
	Status  int    `json:"status"`
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type imageReq struct {
	Image  string `url:"image"`
	Thumbs string `url:"thumbs,omitempty"`
}

type image struct {
	Link    string            `json:"link"`
	Mime    string            `json:"mime"`
	Name    string            `json:"name"`
	Hash    string            `json:"hash"`
	Size    int64             `json:"size"`
	Width   int               `json:"width"`
	Height  int               `json:"height"`
	OCRText string            `json:"ocrtext"`
	Thumbs  map[string]string `json:"thumbs"`
	UserID  string            `json:"user_id"`
}

// Uploader is used to upload images
type Uploader struct {
	mandibleURL string
	thumbs      thumbnailRequest
}

// New uploader
func New(mandibleURL string, thumbs ...Thumbnail) *Uploader {
	uploader := &Uploader{
		mandibleURL: mandibleURL,
		thumbs:      thumbnailRequest{},
	}

	for i := range thumbs {
		thumb := thumbs[i]
		uploader.thumbs[thumb.Name] = &thumb
	}

	return uploader
}

func (u *Uploader) doPostRequest(imageURL string) (*image, error) {
	// generate thumbnails
	thumbsJSON, err := json.Marshal(u.thumbs)
	if err != nil {
		return nil, err
	}

	result := imageResp{}

	_, err = sling.New().
		Base(u.mandibleURL).
		Post("/url").
		BodyForm(&imageReq{
			Image:  imageURL,
			Thumbs: string(thumbsJSON),
		}).
		Receive(&result, &result)

	if err != nil {
		return nil, err
	}

	// check result
	if !result.Success {
		return nil, fmt.Errorf("Unsuccessfull upload: server returned status %v; error: %v", result.Status, result.Error)
	}

	return result.Data, nil
}

// UploadImageByURL is used to upload image to Mandible by it's URL
func (u *Uploader) UploadImageByURL(imageURL string) (string, map[string]string, error) {

	data, err := u.doPostRequest(imageURL)
	if err != nil {
		return "", nil, err
	}

	return data.Link, data.Thumbs, nil
}

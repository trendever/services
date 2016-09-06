package saver

import (
	"savetrend/conf"
	"savetrend/tumbmap"

	"proto/core"
	"utils/mandible"
)

// image uploaders
var (
	mandibleURL = conf.GetSettings().MandibleURL

	productUploader *mandible.Uploader
	avatarUploader  = mandible.New(mandibleURL)
)

func init() {
	thumbnails := []mandible.Thumbnail{}
	for name, info := range tumbmap.ThumbByName {
		thumbnails = append(thumbnails, mandible.Thumbnail{
			Name:   name,
			Width:  info.Size,
			Height: info.Size,
			Shape:  info.Shape,
		})
	}
	mandible.New(mandibleURL, thumbnails...)
}

func generateThumbnails(imageURL string) ([]*core.ImageCandidate, error) {

	url, thumbs, err := productUploader.UploadImageByURL(imageURL)
	if err != nil {
		return nil, err
	}

	out := []*core.ImageCandidate{
		{Name: "Max", Url: url},
	}

	for k, v := range thumbs {

		var thumbName, imageURL = k, v
		out = append(out, &core.ImageCandidate{
			Name: thumbName,
			Url:  imageURL,
		})
	}

	return out, nil

}

func uploadAvatar(url string) (string, error) {
	url, _, err := avatarUploader.UploadImageByURL(url)
	return url, err
}

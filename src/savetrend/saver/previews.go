package saver

import (
	"savetrend/conf"

	"proto/core"
	"utils/mandible"
)

var thumbnails = []mandible.Thumbnail{
	{"XL", 1080, 1080, "thumb"},
	{"L", 750, 750, "thumb"},
	{"M_square", 480, 480, "square"},
	{"S_square", 306, 306, "square"},
}

// image uploaders
var (
	mandibleURL = conf.GetSettings().MandibleURL

	productUploader = mandible.New(mandibleURL, thumbnails...)
	avatarUploader  = mandible.New(mandibleURL)
)

func generateThumbnails(imageURL string) ([]*core.ImageCandidate, error) {

	url, thumbs, err := productUploader.UploadImageByURL(imageURL)
	if err != nil {
		return nil, err
	}

	out := []*core.ImageCandidate{
		&core.ImageCandidate{Name: "Max", Url: url},
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

package imageprocessor

import (
	"mandible/imageprocessor/processorcommand"
	"mandible/uploadedfile"
)

type ExifStripper struct{}

func (this *ExifStripper) Process(image *uploadedfile.UploadedFile) error {
	if !image.IsJpeg() {
		return nil
	}

	err := processorcommand.StripMetadata(image.GetPath())
	if err != nil {
		return err
	}

	return nil
}

func (this *ExifStripper) String() string {
	return "EXIF stripper"
}

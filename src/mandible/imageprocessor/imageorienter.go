package imageprocessor

import (
	"mandible/imageprocessor/processorcommand"
	"mandible/uploadedfile"
)

type ImageOrienter struct{}

func (this *ImageOrienter) Process(image *uploadedfile.UploadedFile) error {
	filename, err := processorcommand.FixOrientation(image.GetPath())
	if err != nil {
		return err
	}

	image.SetPath(filename)

	return nil
}

func (this *ImageOrienter) String() string {
	return "Image orienter"
}

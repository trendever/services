package uploadedfile

import (
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"

	"mandible/imageprocessor/processorcommand"
	"mandible/imageprocessor/thumbType"
)

var (
	defaultQuality   = 83
	maxImageSideSize = 10000
)

type ThumbFile struct {
	localPath string

	Name          string
	Width         int
	MaxWidth      int
	Height        int
	MaxHeight     int
	Shape         string
	CropGravity   string
	CropWidth     int
	CropHeight    int
	CropRatio     string
	Quality       int
	StoreURI      string
	DesiredFormat string
	NoStore       bool
}

func NewThumbFile(width, maxWidth, height, maxHeight int, name, shape, path, cropGravity string, cropWidth, cropHeight int, cropRatio string, quality int, desiredFormat string, noStore bool) *ThumbFile {
	if quality == 0 {
		quality = defaultQuality
	}

	return &ThumbFile{
		localPath: path,

		Name:          name,
		Width:         width,
		MaxWidth:      maxWidth,
		Height:        height,
		MaxHeight:     maxHeight,
		Shape:         shape,
		CropGravity:   cropGravity,
		CropWidth:     cropWidth,
		CropHeight:    cropHeight,
		CropRatio:     cropRatio,
		Quality:       quality,
		StoreURI:      "",
		DesiredFormat: desiredFormat,
		NoStore:       noStore,
	}
}

func (this *ThumbFile) GetNoStore() bool {
	return this.NoStore
}

func (this *ThumbFile) SetPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Error when creating thumbnail %s", this.Name))
	}

	this.localPath = path

	return nil
}

func (this *ThumbFile) GetPath() string {
	return this.localPath
}

func (this *ThumbFile) GetOutputFormat(original *UploadedFile) thumbType.ThumbType {
	if this.DesiredFormat != "" {
		return thumbType.FromString(this.DesiredFormat)
	}

	return thumbType.FromMime(original.GetMime())
}

func (this *ThumbFile) ComputeWidth(original *UploadedFile) int {
	width := this.Width

	oWidth, _, err := original.Dimensions()
	if err != nil {
		return 0
	}

	if this.MaxWidth > 0 {
		width = int(math.Min(float64(oWidth), float64(this.MaxWidth)))
	}

	return width
}

func (this *ThumbFile) ComputeHeight(original *UploadedFile) int {
	height := this.Height

	_, oHeight, err := original.Dimensions()
	if err != nil {
		return 0
	}

	if this.MaxHeight > 0 {
		height = int(math.Min(float64(oHeight), float64(this.MaxHeight)))
	}

	return height
}

func (this *ThumbFile) ComputeCrop(original *UploadedFile) (int, int, error) {
	re := regexp.MustCompile("(.*):(.*)")
	matches := re.FindStringSubmatch(this.CropRatio)
	if len(matches) != 3 {
		return 0, 0, errors.New("Invalid crop_ratio")
	}

	wRatio, werr := strconv.ParseFloat(matches[1], 64)
	hRatio, herr := strconv.ParseFloat(matches[2], 64)
	if werr != nil || herr != nil {
		return 0, 0, errors.New("Invalid crop_ratio")
	}

	var cropWidth, cropHeight float64

	if wRatio >= hRatio {
		wRatio = wRatio / hRatio
		hRatio = 1
		cropWidth = math.Ceil(float64(this.ComputeHeight(original)) * wRatio)
		cropHeight = math.Ceil(float64(this.ComputeHeight(original)) * hRatio)
	} else {
		hRatio = hRatio / wRatio
		wRatio = 1
		cropWidth = math.Ceil(float64(this.ComputeWidth(original)) * wRatio)
		cropHeight = math.Ceil(float64(this.ComputeWidth(original)) * hRatio)
	}

	return int(cropWidth), int(cropHeight), nil
}

func (this *ThumbFile) Process(original *UploadedFile) error {
	switch this.Shape {
	case "circle":
		return this.processCircle(original)
	case "thumb":
		return this.processThumb(original)
	case "square":
		return this.processSquare(original)
	case "custom":
		return this.processCustom(original)
	case "instagram":
		return this.processInstagram(original)
	default:
		return this.processFull(original)
	}
}

func (this *ThumbFile) String() string {
	return fmt.Sprintf("Thumbnail of <%s>", this.Name)
}

func (this *ThumbFile) processSquare(original *UploadedFile) error {
	if this.Width == 0 {
		return errors.New("Width cannot be 0")
	}
	if this.Width > maxImageSideSize {
		return errors.New("Width too large")
	}

	filename, err := processorcommand.SquareThumb(original.GetPath(), this.Name, this.Width, this.Quality, this.GetOutputFormat(original))
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}

func (this *ThumbFile) processInstagram(original *UploadedFile) error {
	width, height, err := original.Dimensions()
	if err != nil {
		return err
	}
	ratio := float32(width) / float32(height)
	switch {
	case ratio >= 4.0/3 && ratio <= 16.0/10:
		this.Width = 1080
		this.Height = 1350
		return this.processSquare(original)
	case ratio < 4.0/3:
		if height > 1350 {
			this.Height = 1350
		} else {
			this.Height = height
		}
		this.Width = int(float32(this.Height)/4*3) + 1
	case ratio > 16.0/10:
		if width > 1050 {
			this.Width = 1050
		} else {
			this.Width = width
		}
		this.Height = int(float32(this.Width)*10/16) + 1
	}

	filename, err := processorcommand.ExtentThumb(original.GetPath(), this.Name, this.Width, this.Height, "white", this.Quality, thumbType.JPG)
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}

func (this *ThumbFile) processCircle(original *UploadedFile) error {
	if this.Width == 0 {
		return errors.New("Width cannot be 0")
	}
	if this.Width > maxImageSideSize {
		return errors.New("Width too large")
	}

	//Circle thumbs should always be PNGs
	outputFormat := thumbType.FromString("png")

	filename, err := processorcommand.CircleThumb(original.GetPath(), this.Name, this.Width, this.Quality, outputFormat)
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}

func (this *ThumbFile) processThumb(original *UploadedFile) error {
	if this.Width == 0 {
		return errors.New("Width cannot be 0")
	}
	if this.Width > maxImageSideSize {
		return errors.New("Width too large")
	}
	if this.Height == 0 {
		return errors.New("Height cannot be 0")
	}
	if this.Height > maxImageSideSize {
		return errors.New("Height too large")
	}

	filename, err := processorcommand.Thumb(original.GetPath(), this.Name, this.Width, this.Height, this.Quality, this.GetOutputFormat(original))
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}

func (this *ThumbFile) processCustom(original *UploadedFile) error {
	cropWidth := this.CropWidth
	cropHeight := this.CropHeight
	var err error

	if this.CropRatio != "" {
		cropWidth, cropHeight, err = this.ComputeCrop(original)
		if err != nil {
			return err
		}
	}

	width := this.ComputeWidth(original)
	height := this.ComputeHeight(original)
	validWidth := width > 0 && width <= maxImageSideSize
	validHeight := height > 0 && height <= maxImageSideSize

	if !validWidth && !validHeight {
		if !validWidth {
			return errors.New("Invalid width")
		}

		return errors.New("Invalid height")
	}

	filename, err := processorcommand.CustomThumb(original.GetPath(), this.Name, width, height, this.CropGravity, cropWidth, cropHeight, this.Quality, this.GetOutputFormat(original))
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}

func (this *ThumbFile) processFull(original *UploadedFile) error {
	filename, err := processorcommand.Full(original.GetPath(), this.Name, this.Quality, this.GetOutputFormat(original))
	if err != nil {
		return err
	}

	if err := this.SetPath(filename); err != nil {
		return err
	}

	return nil
}

package upload

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"time"

	"github.com/disintegration/imaging"
	"github.com/go-resty/resty/v2"
)

const (
	MaxImageSize   = 32 * 1024 * 1024 // 32MB
	MaxResolution  = 2048
	UploadEndpoint = "/prints"
	
	// Standard resolutions for VRChat prints
	Print1080pWidth  = 1920
	Print1080pHeight = 1080
)

type Uploader struct {
	client *resty.Client
}

type Options struct {
	ImagePath string
	Note      string
	WorldID   string
	WorldName string
	NoResize  bool
}

type UploadResult struct {
	FileID     string    `json:"fileId"`
	AuthorID   string    `json:"authorId"`
	AuthorName string    `json:"authorName"`
	CreatedAt  time.Time `json:"createdAt"`
	WorldID    string    `json:"worldId"`
	WorldName  string    `json:"worldName"`
}

func New(client *resty.Client) *Uploader {
	return &Uploader{
		client: client,
	}
}

// createMultipartForm creates a multipart form with image data and metadata
func (u *Uploader) createMultipartForm(imageData []byte, opts Options) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add image file with explicit content type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, filepath.Base(opts.ImagePath)))
	h.Set("Content-Type", "image/png")
	
	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create form part: %w", err)
	}
	
	if _, err := io.Copy(part, bytes.NewReader(imageData)); err != nil {
		return nil, "", fmt.Errorf("failed to write image data: %w", err)
	}

	// Add timestamp
	if err := writer.WriteField("timestamp", time.Now().Format(time.RFC3339)); err != nil {
		return nil, "", fmt.Errorf("failed to write timestamp: %w", err)
	}

	// Add optional fields
	if opts.Note != "" {
		if err := writer.WriteField("note", opts.Note); err != nil {
			return nil, "", fmt.Errorf("failed to write note: %w", err)
		}
	}

	if opts.WorldID != "" {
		if err := writer.WriteField("worldId", opts.WorldID); err != nil {
			return nil, "", fmt.Errorf("failed to write worldId: %w", err)
		}
	}

	if opts.WorldName != "" {
		if err := writer.WriteField("worldName", opts.WorldName); err != nil {
			return nil, "", fmt.Errorf("failed to write worldName: %w", err)
		}
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return body, contentType, nil
}

// resizeImage resizes the image according to the specified options
func resizeImage(img image.Image, noResize bool) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if !noResize {
		// Resize to 1080p for prints (as per VRChat spec)
		// 1920x1080 or 1080x1920 depending on orientation
		if width > height {
			return imaging.Resize(img, Print1080pWidth, Print1080pHeight, imaging.Lanczos)
		} else {
			return imaging.Resize(img, Print1080pHeight, Print1080pWidth, imaging.Lanczos)
		}
	} else {
		// Keep original resolution when noResize is true, but limit to 2048x2048
		if width > MaxResolution || height > MaxResolution {
			// Resize to fit within MaxResolution while maintaining aspect ratio
			if width > height {
				return imaging.Resize(img, MaxResolution, 0, imaging.Lanczos)
			} else {
				return imaging.Resize(img, 0, MaxResolution, imaging.Lanczos)
			}
		}
		// Return original image if no resize needed
		return img
	}
}

func (u *Uploader) Upload(opts Options) (*UploadResult, error) {
	// Validate and prepare image
	imageData, err := u.prepareImage(opts.ImagePath, opts.NoResize)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare image: %w", err)
	}

	// Create multipart form
	body, contentType, err := u.createMultipartForm(imageData, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart form: %w", err)
	}

	// Upload
	resp, err := u.client.R().
		SetHeader("Content-Type", contentType).
		SetBody(body.Bytes()).
		SetResult(&UploadResult{}).
		Post(UploadEndpoint)

	if err != nil {
		return nil, fmt.Errorf("upload request failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	return resp.Result().(*UploadResult), nil
}

func (u *Uploader) prepareImage(imagePath string, noResize bool) ([]byte, error) {
	// Check file exists
	info, err := os.Stat(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat image file: %w", err)
	}

	// Check file size
	if info.Size() > MaxImageSize {
		return nil, fmt.Errorf("image file too large: %d bytes (max: %d bytes)", info.Size(), MaxImageSize)
	}

	// Open and decode image
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize image according to options
	img = resizeImage(img, noResize)

	// Encode as PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode image as PNG: %w", err)
	}

	// Check final size
	if buf.Len() > MaxImageSize {
		return nil, fmt.Errorf("encoded image too large: %d bytes (max: %d bytes)", buf.Len(), MaxImageSize)
	}

	// Image prepared and converted to PNG
	
	return buf.Bytes(), nil
}
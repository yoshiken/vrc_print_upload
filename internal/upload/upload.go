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

func (u *Uploader) Upload(opts Options) (*UploadResult, error) {
	// Validate and prepare image
	imageData, err := u.prepareImage(opts.ImagePath, opts.NoResize)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare image: %w", err)
	}

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add image file with explicit content type
	fmt.Printf("Creating form file with name: image, filename: %s\n", filepath.Base(opts.ImagePath))
	
	// Create form field with explicit headers
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, filepath.Base(opts.ImagePath)))
	h.Set("Content-Type", "image/png")
	
	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, fmt.Errorf("failed to create form part: %w", err)
	}
	
	if _, err := io.Copy(part, bytes.NewReader(imageData)); err != nil {
		return nil, fmt.Errorf("failed to write image data: %w", err)
	}

	// Add timestamp
	if err := writer.WriteField("timestamp", time.Now().Format(time.RFC3339)); err != nil {
		return nil, fmt.Errorf("failed to write timestamp: %w", err)
	}

	// Add optional fields
	if opts.Note != "" {
		if err := writer.WriteField("note", opts.Note); err != nil {
			return nil, fmt.Errorf("failed to write note: %w", err)
		}
	}

	if opts.WorldID != "" {
		if err := writer.WriteField("worldId", opts.WorldID); err != nil {
			return nil, fmt.Errorf("failed to write worldId: %w", err)
		}
	}

	if opts.WorldName != "" {
		if err := writer.WriteField("worldName", opts.WorldName); err != nil {
			return nil, fmt.Errorf("failed to write worldName: %w", err)
		}
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Upload
	fmt.Printf("Uploading to: %s%s\n", u.client.BaseURL, UploadEndpoint)
	fmt.Printf("Content-Type: %s\n", contentType)
	fmt.Printf("Body size: %d bytes\n", body.Len())
	
	// Debug: show first 500 bytes of the request body
	bodyBytes := body.Bytes()
	if len(bodyBytes) > 500 {
		fmt.Printf("First 500 bytes of body:\n%s\n", string(bodyBytes[:500]))
	}
	
	resp, err := u.client.R().
		SetHeader("Content-Type", contentType).
		SetBody(bodyBytes).
		SetResult(&UploadResult{}).
		Post(UploadEndpoint)

	if err != nil {
		return nil, fmt.Errorf("upload request failed: %w", err)
	}

	fmt.Printf("Response status: %d\n", resp.StatusCode())
	fmt.Printf("Response headers: %v\n", resp.Header())
	
	if resp.StatusCode() != 200 {
		fmt.Printf("Response body: %s\n", resp.String())
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

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Check if resizing is needed
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width > MaxResolution || height > MaxResolution {
		// Resize to fit within MaxResolution while maintaining aspect ratio
		if width > height {
			img = imaging.Resize(img, MaxResolution, 0, imaging.Lanczos)
		} else {
			img = imaging.Resize(img, 0, MaxResolution, imaging.Lanczos)
		}
		fmt.Printf("Image resized to fit within %dx%d\n", MaxResolution, MaxResolution)
	}

	// Convert to 1080p for prints (as per VRChat spec)
	// 1920x1080 or 1080x1920 depending on orientation
	// Skip this step if noResize is true
	if !noResize {
		if width > height {
			img = imaging.Resize(img, 1920, 1080, imaging.Lanczos)
		} else {
			img = imaging.Resize(img, 1080, 1920, imaging.Lanczos)
		}
		fmt.Printf("Image resized to 1080p\n")
	} else {
		fmt.Printf("Keeping original resolution (up to %dx%d)\n", bounds.Dx(), bounds.Dy())
	}

	// Encode as PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode image as PNG: %w", err)
	}

	// Check final size
	if buf.Len() > MaxImageSize {
		return nil, fmt.Errorf("encoded image too large: %d bytes (max: %d bytes)", buf.Len(), MaxImageSize)
	}

	fmt.Printf("Image prepared: %s format, converted to PNG (%d bytes)\n", format, buf.Len())
	
	return buf.Bytes(), nil
}
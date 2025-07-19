package upload

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareImage(t *testing.T) {
	// Create temp directory for test images
	tempDir, err := os.MkdirTemp("", "vrc-print-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name          string
		createImage   func(string) error
		noResize      bool
		expectError   bool
		expectedMsg   string
		checkResult   func(*testing.T, []byte)
	}{
		{
			name: "PNG image - no resize",
			createImage: func(path string) error {
				return createTestImage(path, "png", 1000, 800)
			},
			noResize: true,
			checkResult: func(t *testing.T, data []byte) {
				// Verify it's a valid PNG
				img, format, err := image.Decode(bytes.NewReader(data))
				require.NoError(t, err)
				assert.Equal(t, "png", format)
				assert.Equal(t, 1000, img.Bounds().Dx())
				assert.Equal(t, 800, img.Bounds().Dy())
			},
		},
		{
			name: "PNG image - with resize",
			createImage: func(path string) error {
				return createTestImage(path, "png", 1000, 800)
			},
			noResize: false,
			checkResult: func(t *testing.T, data []byte) {
				img, format, err := image.Decode(bytes.NewReader(data))
				require.NoError(t, err)
				assert.Equal(t, "png", format)
				// Should be resized to 1920x1080 (landscape)
				assert.Equal(t, 1920, img.Bounds().Dx())
				assert.Equal(t, 1080, img.Bounds().Dy())
			},
		},
		{
			name: "JPEG image - converts to PNG",
			createImage: func(path string) error {
				return createTestImage(path, "jpeg", 800, 600)
			},
			noResize: false,
			checkResult: func(t *testing.T, data []byte) {
				img, format, err := image.Decode(bytes.NewReader(data))
				require.NoError(t, err)
				assert.Equal(t, "png", format)
				assert.Equal(t, 1920, img.Bounds().Dx())
				assert.Equal(t, 1080, img.Bounds().Dy())
			},
		},
		{
			name: "Portrait image - with resize",
			createImage: func(path string) error {
				return createTestImage(path, "png", 800, 1200)
			},
			noResize: false,
			checkResult: func(t *testing.T, data []byte) {
				img, _, err := image.Decode(bytes.NewReader(data))
				require.NoError(t, err)
				// Should be resized to 1080x1920 (portrait)
				assert.Equal(t, 1080, img.Bounds().Dx())
				assert.Equal(t, 1920, img.Bounds().Dy())
			},
		},
		{
			name: "Large image - auto resize to max resolution",
			createImage: func(path string) error {
				return createTestImage(path, "png", 3000, 4000)
			},
			noResize: true,
			checkResult: func(t *testing.T, data []byte) {
				img, _, err := image.Decode(bytes.NewReader(data))
				require.NoError(t, err)
				// Should be resized to fit within 2048x2048
				assert.LessOrEqual(t, img.Bounds().Dx(), MaxResolution)
				assert.LessOrEqual(t, img.Bounds().Dy(), MaxResolution)
				// Should maintain aspect ratio
				assert.Equal(t, 1536, img.Bounds().Dx()) // 2048 * (3000/4000)
				assert.Equal(t, 2048, img.Bounds().Dy())
			},
		},
		{
			name: "GIF image - converts to PNG",
			createImage: func(path string) error {
				return createTestImage(path, "gif", 500, 500)
			},
			noResize: true,
			checkResult: func(t *testing.T, data []byte) {
				img, format, err := image.Decode(bytes.NewReader(data))
				require.NoError(t, err)
				assert.Equal(t, "png", format)
				assert.Equal(t, 500, img.Bounds().Dx())
				assert.Equal(t, 500, img.Bounds().Dy())
			},
		},
		{
			name: "Non-existent file",
			createImage: func(path string) error {
				// Remove the file if it exists
				os.Remove(path)
				return nil
			},
			noResize:    false,
			expectError: true,
			expectedMsg: "failed to stat image file",
		},
		{
			name: "Invalid image file",
			createImage: func(path string) error {
				return os.WriteFile(path, []byte("not an image"), 0644)
			},
			expectError: true,
			expectedMsg: "failed to decode image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imagePath := filepath.Join(tempDir, "test.img")
			if tt.createImage != nil {
				err := tt.createImage(imagePath)
				require.NoError(t, err)
			}

			uploader := &Uploader{}
			data, err := uploader.prepareImage(imagePath, tt.noResize)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, data)
				if tt.checkResult != nil {
					tt.checkResult(t, data)
				}
			}
		})
	}
}

func TestUpload(t *testing.T) {
	// Create temp directory for test images
	tempDir, err := os.MkdirTemp("", "vrc-print-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test image
	imagePath := filepath.Join(tempDir, "test.png")
	err = createTestImage(imagePath, "png", 100, 100)
	require.NoError(t, err)

	// Create HTTP client with mock
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Mock successful upload response
	mockResponse := &UploadResult{
		FileID:     "file_12345",
		AuthorID:   "usr_12345",
		AuthorName: "TestUser",
		CreatedAt:  time.Now(),
		WorldID:    "wrld_12345",
		WorldName:  "Test World",
	}

	httpmock.RegisterResponder("POST", "https://api.vrchat.cloud/api/1/prints",
		func(req *http.Request) (*http.Response, error) {
			// Verify multipart form data
			err := req.ParseMultipartForm(32 << 20)
			if err != nil {
				return httpmock.NewStringResponse(400, "Invalid multipart form"), nil
			}

			// Check required fields
			if req.MultipartForm.File["image"] == nil {
				return httpmock.NewStringResponse(400, "Missing image file"), nil
			}

			// Return success response
			resp, _ := httpmock.NewJsonResponse(200, mockResponse)
			return resp, nil
		})

	// Set base URL
	client.SetBaseURL("https://api.vrchat.cloud/api/1")

	uploader := New(client)

	tests := []struct {
		name        string
		opts        Options
		expectError bool
	}{
		{
			name: "Basic upload",
			opts: Options{
				ImagePath: imagePath,
			},
		},
		{
			name: "Upload with metadata",
			opts: Options{
				ImagePath: imagePath,
				Note:      "Test note",
				WorldID:   "wrld_12345",
				WorldName: "Test World",
			},
		},
		{
			name: "Upload with no-resize",
			opts: Options{
				ImagePath: imagePath,
				NoResize:  true,
			},
		},
		{
			name: "Invalid image path",
			opts: Options{
				ImagePath: "/non/existent/image.png",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := uploader.Upload(tt.opts)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, mockResponse.FileID, result.FileID)
				assert.Equal(t, mockResponse.AuthorID, result.AuthorID)
			}
		})
	}
}

func TestUploadErrorResponses(t *testing.T) {
	// Create temp directory for test images
	tempDir, err := os.MkdirTemp("", "vrc-print-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test image
	imagePath := filepath.Join(tempDir, "test.png")
	err = createTestImage(imagePath, "png", 100, 100)
	require.NoError(t, err)

	tests := []struct {
		name           string
		mockStatusCode int
		mockResponse   string
		expectedError  string
	}{
		{
			name:           "Unauthorized",
			mockStatusCode: 401,
			mockResponse:   `{"error": "Unauthorized"}`,
			expectedError:  "upload failed with status 401",
		},
		{
			name:           "Rate limited",
			mockStatusCode: 429,
			mockResponse:   `{"error": "Rate limit exceeded"}`,
			expectedError:  "upload failed with status 429",
		},
		{
			name:           "Server error",
			mockStatusCode: 500,
			mockResponse:   `{"error": "Internal server error"}`,
			expectedError:  "upload failed with status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP client with mock
			client := resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			defer httpmock.DeactivateAndReset()

			// Mock error response
			httpmock.RegisterResponder("POST", "https://api.vrchat.cloud/api/1/prints",
				httpmock.NewStringResponder(tt.mockStatusCode, tt.mockResponse))

			client.SetBaseURL("https://api.vrchat.cloud/api/1")
			uploader := New(client)

			result, err := uploader.Upload(Options{
				ImagePath: imagePath,
			})

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// Helper function to create test images
func createTestImage(path string, format string, width, height int) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Fill with a gradient for visual verification
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x * 255 / width),
				G: uint8(y * 255 / height),
				B: 128,
				A: 255,
			})
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	switch format {
	case "png":
		return png.Encode(file, img)
	case "jpeg", "jpg":
		return jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	case "gif":
		return gif.Encode(file, img, &gif.Options{})
	default:
		return png.Encode(file, img)
	}
}
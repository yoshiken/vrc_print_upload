package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yoshiken/vrc-print-upload/internal/auth"
	"github.com/yoshiken/vrc-print-upload/internal/config"
	"github.com/yoshiken/vrc-print-upload/internal/upload"
)

// App struct
type App struct {
	ctx           context.Context
	config        *config.Config
	authClient    *auth.Client
	uploadService *upload.Uploader
}

// LoginRequest represents login request data
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents login response data
type LoginResponse struct {
	Success           bool     `json:"success"`
	Message           string   `json:"message"`
	RequiresTwoFactor bool     `json:"requiresTwoFactor"`
	UserDisplayName   string   `json:"userDisplayName,omitempty"`
	Errors            []string `json:"errors,omitempty"`
}

// TwoFactorRequest represents 2FA request data
type TwoFactorRequest struct {
	Code          string `json:"code"`
	IsRecoveryCode bool   `json:"isRecoveryCode"`
}

// UploadRequest represents upload request data
type UploadRequest struct {
	ImagePath string `json:"imagePath"`
	Note      string `json:"note"`
	WorldID   string `json:"worldId"`
	WorldName string `json:"worldName"`
	NoResize  bool   `json:"noResize"`
}

// UploadResponse represents upload response data
type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	FileID  string `json:"fileId,omitempty"`
	Error   string `json:"error,omitempty"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize auth client
	authClient := auth.NewClient(cfg)

	return &App{
		config:     cfg,
		authClient: authClient,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// Initialize upload service if user is already authenticated
	if a.IsAuthenticated() {
		a.uploadService = upload.New(a.authClient.GetHTTPClient())
	}
}

// IsAuthenticated checks if user is logged in
func (a *App) IsAuthenticated() bool {
	// First check if we have valid cookies
	if !a.authClient.IsAuthenticated() {
		return false
	}
	
	// Also verify with the API to make sure the session is still valid
	_, err := a.authClient.GetCurrentUser()
	return err == nil
}

// Login attempts to log in the user
func (a *App) Login(req LoginRequest) LoginResponse {
	opts := auth.LoginOptions{
		Username: req.Username,
		Password: req.Password,
	}

	err := a.authClient.Login(opts)
	if err != nil {
		// Check if it's a 2FA error
		errMsg := err.Error()
		if strings.Contains(errMsg, "2FA") || strings.Contains(errMsg, "two-factor") {
			return LoginResponse{
				Success:           false,
				RequiresTwoFactor: true,
				Message:           "Two-factor authentication required",
			}
		}

		return LoginResponse{
			Success: false,
			Message: fmt.Sprintf("Login failed: %v", err),
		}
	}

	// Get user info after successful login
	user, err := a.authClient.GetCurrentUser()
	displayName := ""
	if err == nil && user != nil {
		displayName = user.DisplayName
	}

	// Initialize upload service after successful login
	a.uploadService = upload.New(a.authClient.GetHTTPClient())

	return LoginResponse{
		Success:         true,
		Message:         "Login successful",
		UserDisplayName: displayName,
	}
}

// VerifyTwoFactor verifies 2FA code
func (a *App) VerifyTwoFactor(req TwoFactorRequest) LoginResponse {
	var err error
	
	if req.IsRecoveryCode {
		err = a.authClient.VerifyRecoveryCode(req.Code)
	} else {
		err = a.authClient.VerifyTOTPCode(req.Code)
	}
	
	if err != nil {
		return LoginResponse{
			Success: false,
			Message: fmt.Sprintf("2FA verification failed: %v", err),
		}
	}
	
	// Get user info after successful 2FA verification
	user, err := a.authClient.GetCurrentUser()
	displayName := ""
	if err == nil && user != nil {
		displayName = user.DisplayName
	}

	// Initialize upload service after successful 2FA
	a.uploadService = upload.New(a.authClient.GetHTTPClient())

	return LoginResponse{
		Success:         true,
		Message:         "2FA verification successful",
		UserDisplayName: displayName,
	}
}

// Logout logs out the user
func (a *App) Logout() LoginResponse {
	err := a.authClient.Logout()
	if err != nil {
		return LoginResponse{
			Success: false,
			Message: fmt.Sprintf("Logout failed: %v", err),
		}
	}

	a.uploadService = nil
	return LoginResponse{
		Success: true,
		Message: "Logged out successfully",
	}
}

// GetCurrentUser returns current user info
func (a *App) GetCurrentUser() LoginResponse {
	user, err := a.authClient.GetCurrentUser()
	if err != nil {
		return LoginResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get user info: %v", err),
		}
	}

	return LoginResponse{
		Success:         true,
		UserDisplayName: user.DisplayName,
	}
}

// UploadImage uploads an image to VRChat
func (a *App) UploadImage(req UploadRequest) UploadResponse {
	if a.uploadService == nil {
		return UploadResponse{
			Success: false,
			Error:   "Not authenticated. Please log in first.",
		}
	}

	// Validate file path
	if req.ImagePath == "" {
		return UploadResponse{
			Success: false,
			Error:   "No image selected",
		}
	}

	// Convert to absolute path if needed
	absPath, err := filepath.Abs(req.ImagePath)
	if err != nil {
		return UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid file path: %v", err),
		}
	}

	opts := upload.Options{
		ImagePath: absPath,
		Note:      req.Note,
		WorldID:   req.WorldID,
		WorldName: req.WorldName,
		NoResize:  req.NoResize,
	}

	result, err := a.uploadService.Upload(opts)
	if err != nil {
		return UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Upload failed: %v", err),
		}
	}

	return UploadResponse{
		Success: true,
		Message: "Upload successful",
		FileID:  result.FileID,
	}
}

// OpenFileDialog opens a file dialog and returns the selected file path
func (a *App) OpenFileDialog() (string, error) {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Image File",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Image Files",
				Pattern:     "*.png;*.jpg;*.jpeg;*.gif",
			},
			{
				DisplayName: "PNG Files",
				Pattern:     "*.png",
			},
			{
				DisplayName: "JPEG Files", 
				Pattern:     "*.jpg;*.jpeg",
			},
			{
				DisplayName: "GIF Files",
				Pattern:     "*.gif",
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}

	return filePath, nil
}

// ValidateImageFile validates if a file is a supported image format
func (a *App) ValidateImageFile(filePath string) map[string]interface{} {
	if filePath == "" {
		return map[string]interface{}{
			"valid": false,
			"error": "No file path provided",
		}
	}

	ext := filepath.Ext(filePath)
	supportedExts := []string{".png", ".jpg", ".jpeg", ".gif"}
	
	for _, supportedExt := range supportedExts {
		if ext == supportedExt {
			return map[string]interface{}{
				"valid": true,
				"type":  ext,
			}
		}
	}

	return map[string]interface{}{
		"valid": false,
		"error": fmt.Sprintf("Unsupported file format: %s", ext),
	}
}


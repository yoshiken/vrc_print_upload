package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultConfig(t *testing.T) {
	// Reset viper to clean state
	viper.Reset()
	
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}
	
	// Set temporary home directory
	os.Setenv("HOME", tempHome)
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		os.Setenv("USERPROFILE", tempHome)
	}
	
	// Restore original home directory after test
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
			if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
				os.Setenv("USERPROFILE", originalHome)
			}
		}
	}()

	cfg, err := Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check default values
	assert.Equal(t, "https://api.vrchat.cloud/api/1", cfg.APIBaseURL)
	assert.Equal(t, filepath.Join(tempHome, ".vrc-print"), cfg.ConfigDir())

	// Check that config directory was created
	_, err = os.Stat(cfg.ConfigDir())
	assert.NoError(t, err)
}

func TestLoad_WithConfigFile(t *testing.T) {
	// Reset viper to clean state
	viper.Reset()
	
	// Create temporary directory
	tempDir := t.TempDir()
	
	// Create config file
	configFile := filepath.Join(tempDir, "test-config.yaml")
	configContent := `api_base_url: "https://custom.api.com/v2"`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}
	
	// Set temporary home directory
	os.Setenv("HOME", tempHome)
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		os.Setenv("USERPROFILE", tempHome)
	}
	
	// Restore original home directory after test
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
			if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
				os.Setenv("USERPROFILE", originalHome)
			}
		}
	}()

	cfg, err := Load(configFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check that custom API URL was loaded
	assert.Equal(t, "https://custom.api.com/v2", cfg.APIBaseURL)
	assert.Equal(t, filepath.Join(tempHome, ".vrc-print"), cfg.ConfigDir())
}

func TestLoad_WithEnvironmentVariable(t *testing.T) {
	// Reset viper to clean state
	viper.Reset()
	
	// Set environment variable
	originalEnv := os.Getenv("VRC_PRINT_API_BASE_URL")
	os.Setenv("VRC_PRINT_API_BASE_URL", "https://env.api.com/v1")
	defer func() {
		if originalEnv != "" {
			os.Setenv("VRC_PRINT_API_BASE_URL", originalEnv)
		} else {
			os.Unsetenv("VRC_PRINT_API_BASE_URL")
		}
	}()

	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}
	
	// Set temporary home directory
	os.Setenv("HOME", tempHome)
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		os.Setenv("USERPROFILE", tempHome)
	}
	
	// Restore original home directory after test
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
			if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
				os.Setenv("USERPROFILE", originalHome)
			}
		}
	}()

	cfg, err := Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check that environment variable overrode default
	assert.Equal(t, "https://env.api.com/v1", cfg.APIBaseURL)
}

func TestLoad_ConfigFileInDefaultLocation(t *testing.T) {
	// Reset viper to clean state
	viper.Reset()
	
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}
	
	// Set temporary home directory
	os.Setenv("HOME", tempHome)
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		os.Setenv("USERPROFILE", tempHome)
	}
	
	// Restore original home directory after test
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
			if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
				os.Setenv("USERPROFILE", originalHome)
			}
		}
	}()

	// Create config directory and file
	configDir := filepath.Join(tempHome, ".vrc-print")
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)

	configFile := filepath.Join(configDir, "config.yaml")
	configContent := `api_base_url: "https://default.config.com"`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err := Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check that config file was loaded from default location
	assert.Equal(t, "https://default.config.com", cfg.APIBaseURL)
}

func TestLoad_InvalidConfigFile(t *testing.T) {
	// Reset viper to clean state
	viper.Reset()
	
	// Create temporary directory
	tempDir := t.TempDir()
	
	// Create invalid config file
	configFile := filepath.Join(tempDir, "invalid-config.yaml")
	configContent := `invalid yaml content: [unclosed bracket`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}
	
	// Set temporary home directory
	os.Setenv("HOME", tempHome)
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		os.Setenv("USERPROFILE", tempHome)
	}
	
	// Restore original home directory after test
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
			if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
				os.Setenv("USERPROFILE", originalHome)
			}
		}
	}()

	cfg, err := Load(configFile)
	
	// Should return error for invalid config file
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to read config")
}

func TestConfigDir(t *testing.T) {
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}
	
	// Set temporary home directory
	os.Setenv("HOME", tempHome)
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		os.Setenv("USERPROFILE", tempHome)
	}
	
	// Restore original home directory after test
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
			if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
				os.Setenv("USERPROFILE", originalHome)
			}
		}
	}()

	cfg := &Config{
		configDir: filepath.Join(tempHome, ".vrc-print"),
	}

	result := cfg.ConfigDir()
	expected := filepath.Join(tempHome, ".vrc-print")
	assert.Equal(t, expected, result)
}

func TestCookieFile(t *testing.T) {
	cfg := &Config{}

	result := cfg.CookieFile()
	
	// Should return a path ending with cookies.json (exe directory)
	assert.True(t, strings.HasSuffix(result, "cookies.json"))
	// Should not contain .vrc-print directory anymore
	assert.False(t, strings.Contains(result, ".vrc-print"))
}

func TestLoad_ConfigDirectoryPermissions(t *testing.T) {
	// Reset viper to clean state
	viper.Reset()
	
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}
	
	// Set temporary home directory
	os.Setenv("HOME", tempHome)
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		os.Setenv("USERPROFILE", tempHome)
	}
	
	// Restore original home directory after test
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
			if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
				os.Setenv("USERPROFILE", originalHome)
			}
		}
	}()

	cfg, err := Load("")
	require.NoError(t, err)

	// Check that config directory exists and has correct permissions
	info, err := os.Stat(cfg.ConfigDir())
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Check permissions (should be 0700 - readable/writable/executable by owner only)
	mode := info.Mode()
	assert.Equal(t, os.FileMode(0700), mode&0777)
}

func TestLoad_ConfigFileNotFound(t *testing.T) {
	// Reset viper to clean state
	viper.Reset()
	
	// Create temporary home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}
	
	// Set temporary home directory
	os.Setenv("HOME", tempHome)
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		os.Setenv("USERPROFILE", tempHome)
	}
	
	// Restore original home directory after test
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
			if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
				os.Setenv("USERPROFILE", originalHome)
			}
		}
	}()

	// Load config without any config file
	cfg, err := Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Should use default values when config file is not found
	assert.Equal(t, "https://api.vrchat.cloud/api/1", cfg.APIBaseURL)
}
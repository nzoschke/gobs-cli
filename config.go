package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigCmd handles configuration management.
type ConfigCmd struct {
	SetPassword SetPasswordCmd `help:"Set password in config files." cmd:""`
}

// SetPasswordCmd sets the OBS password in configuration files.
type SetPasswordCmd struct {
	Password string `arg:"" help:"Password to set for OBS authentication."`
}

// Run executes the set-password command.
func (cmd *SetPasswordCmd) Run(ctx *context) error {
	// Get user config directory
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get user config directory: %w", err)
	}

	// Define config paths
	gobsConfigPath := filepath.Join(userConfigDir, "gobs-cli", "config.env")
	obsConfigPath := filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "obs-studio", "plugin_config", "obs-websocket", "config.json")

	// Write to gobs-cli config file
	if err := writeGobsConfig(gobsConfigPath, cmd.Password); err != nil {
		return fmt.Errorf("failed to write to gobs-cli config: %w", err)
	}

	// Write to OBS WebSocket config file
	if err := writeObsConfig(obsConfigPath, cmd.Password); err != nil {
		return fmt.Errorf("failed to write to OBS config: %w", err)
	}

	fmt.Fprintf(ctx.Out, "Password set in:\n  %s\n  %s\n", gobsConfigPath, obsConfigPath)
	return nil
}

// writeGobsConfig writes the gobs-cli config file with the new password.
func writeGobsConfig(configPath, password string) error {
	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	content := fmt.Sprintf(`OBS_HOST=localhost
OBS_PORT=4455
OBS_PASSWORD=%s
OBS_TIMEOUT=5
`, password)

	return os.WriteFile(configPath, []byte(content), 0600)
}

// writeObsConfig writes the OBS WebSocket config file with the new password.
func writeObsConfig(configPath, password string) error {
	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	config := map[string]interface{}{
		"alerts_enabled":  false,
		"auth_required":   true,
		"first_load":      false,
		"server_enabled":  true,
		"server_password": password,
		"server_port":     4455,
	}

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(configPath, data, 0600)
}
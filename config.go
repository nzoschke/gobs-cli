package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ConfigCmd handles configuration management.
type ConfigCmd struct {
	SetPassword SetPasswordCmd `help:"Set password in config files." cmd:""`
	LaunchObs   LaunchObsCmd   `help:"Launch OBS if not running." cmd:""`
}

// SetPasswordCmd sets the OBS password in configuration files.
type SetPasswordCmd struct {
	Password string `arg:"" help:"Password to set for OBS authentication."`
}

// Run executes the set-password command.
func (cmd *SetPasswordCmd) Run() error {
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

	fmt.Printf("Password set in:\n  %s\n  %s\n", gobsConfigPath, obsConfigPath)
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

// LaunchObsCmd launches OBS if it's not already running.
type LaunchObsCmd struct{}

// Run executes the launch-obs command.
func (cmd *LaunchObsCmd) Run() error {
	// Check if OBS is already running
	if isObsRunning() {
		fmt.Printf("OBS is already running\n")
		return nil
	}

	// Launch OBS
	if err := launchObs(); err != nil {
		return fmt.Errorf("failed to launch OBS: %w", err)
	}

	fmt.Printf("OBS launched\n")
	return nil
}

// isObsRunning checks if OBS is currently running.
func isObsRunning() bool {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pgrep", "-f", "OBS")
	case "windows":
		cmd = exec.Command("tasklist", "/FI", "IMAGENAME eq obs64.exe", "/FO", "CSV")
	default:
		return false
	}
	err := cmd.Run()
	return err == nil
}

// launchObs launches OBS application.
func launchObs() error {
	switch runtime.GOOS {
	case "darwin":
		obsPath := "/Applications/OBS.app"
		if _, err := os.Stat(obsPath); os.IsNotExist(err) {
			return fmt.Errorf("OBS not found at %s", obsPath)
		}
		return exec.Command("open", obsPath).Start()
	case "windows":
		// Try common installation paths
		paths := []string{
			`C:\Program Files\obs-studio\bin\64bit\obs64.exe`,
			`C:\Program Files (x86)\obs-studio\bin\64bit\obs64.exe`,
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return exec.Command(path).Start()
			}
		}
		return fmt.Errorf("OBS not found in common installation paths")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
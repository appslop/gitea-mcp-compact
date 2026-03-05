package settings

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gitea.com/gitea/gitea-mcp/pkg/log"
)

// Settings represents the configuration stored in the settings file
type Settings struct {
	Token            string `json:"agent_token,omitempty"`
	UserToken        string `json:"user_token,omitempty"`
	Host             string `json:"host,omitempty"`
	CreateRepoAsUser bool   `json:"create_repo_as_user,omitempty"`
	TruncateCompact  int    `json:"truncate_compact,omitempty"`
	TruncateFull     int    `json:"truncate_full,omitempty"`
}

// getConfigPath returns the path to the settings file
// On Windows: C:\Users\<username>\.gitea\mcp\settings.json
// On Unix: ~/.gitea-mcp/settings.json
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Debugf("Could not get user home directory: %v", err)
		return ""
	}
	return filepath.Join(homeDir, ".gitea-mcp", "settings.json")
}

// Load reads the settings file and returns the Settings struct
// Returns an empty Settings struct if the file doesn't exist or cannot be read
func Load() Settings {
	configPath := getConfigPath()
	if configPath == "" {
		return Settings{}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Debugf("Could not read settings file %s: %v", configPath, err)
		}
		return Settings{}
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		log.Debugf("Could not parse settings file %s: %v", configPath, err)
		return Settings{}
	}

	log.Debugf("Loaded settings from %s", configPath)
	return settings
}

// Save writes the settings to the settings file
// Creates the directory if it doesn't exist
func Save(settings Settings) error {
	configPath := getConfigPath()
	if configPath == "" {
		return os.ErrNotExist
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return err
	}

	log.Infof("Settings saved to %s", configPath)
	return nil
}

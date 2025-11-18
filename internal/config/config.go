package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// VCenterConfig contains vCenter connection information
type VCenterConfig struct {
	Host     string `json:"vcenter_url"`
	Username string `json:"username"`
	Password string `json:"-"` // Never store password in file
	Session  string `json:"-"`
	Mode     string `json:"mode,omitempty"`     // "sspi" or "password"
	Insecure bool   `json:"insecure,omitempty"` // allow unsigned certificates
}

// DefaultsConfig for the "defaults" section
type DefaultsConfig struct {
	SnapshotNamePrefix   string `json:"snapshot_name_prefix"`
	IsoDatastorePath     string `json:"iso_datastore_path"`
	SkipMemoryInSnapshot bool   `json:"skip_memory_in_snapshot"`
	GuestUsername        string `json:"guest_username,omitempty"`
}

// UpgradeConfig for the "upgrade" section
type UpgradeConfig struct {
	Parallel       int  `json:"parallel"`
	Reboot         bool `json:"reboot"`
	TimeoutMinutes int  `json:"timeout_minutes"`
	PrecheckDiskGB int  `json:"precheck_disk_gb"`
}

// TimeoutConfig contains detailed timeout settings
type TimeoutConfig struct {
	SignalScriptSeconds int `json:"signal_script_seconds"`
	SignalFilesMinutes  int `json:"signal_files_minutes"`
	TargetOSMinutes     int `json:"target_os_minutes"`
	PowerOffMinutes     int `json:"poweroff_minutes"`
}

// LoggingConfig for the "logging" section
type LoggingConfig struct {
	Level string `json:"level"`
	File  string `json:"file"`
}

// UIConfig for the "ui" section
type UIConfig struct {
	Language string `json:"language"`
	DarkMode bool   `json:"dark_mode"`
}

// AppConfig represents the configuration file structure
type AppConfig struct {
	VCenter  VCenterConfig  `json:"vcenter"`
	Defaults DefaultsConfig `json:"defaults"`
	Upgrade  UpgradeConfig  `json:"upgrade"`
	Timeouts TimeoutConfig  `json:"timeouts"`
	Logging  LoggingConfig  `json:"logging"`
	UI       UIConfig       `json:"ui"`
}

const configFileName = "conf.json"

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("kunde inte hitta home directory: %w", err)
	}
	return filepath.Join(homeDir, configFileName), nil
}

// Load reads the configuration from file
func Load() (*AppConfig, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default configuration
			cfg := createDefaultConfig()
			if saveErr := Save(cfg); saveErr != nil {
				return nil, fmt.Errorf("kunde inte skapa defaultkonfiguration: %w", saveErr)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("kunde inte läsa config: %w", err)
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("kunde inte parsa config: %w", err)
	}

	cfg.applyTimeoutDefaults()

	return &cfg, nil
}

// Save saves the configuration to file
func Save(cfg *AppConfig) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("kunde inte serialisera config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("kunde inte skriva config: %w", err)
	}

	return nil
}

// createDefaultConfig creates a default configuration
func createDefaultConfig() *AppConfig {
	return &AppConfig{
		VCenter: VCenterConfig{
			Host:     "vcenter.example.local",
			Username: "administrator@vsphere.local",
			Mode:     "password",
			Insecure: true,
		},
		Defaults: DefaultsConfig{
			SnapshotNamePrefix:   "pre-upgrade",
			IsoDatastorePath:     "[datastore1] iso/windows-server-2022.iso",
			SkipMemoryInSnapshot: true,
			GuestUsername:        "Administrator",
		},
		Upgrade: UpgradeConfig{
			Parallel:       10, // Number of parallel upgrades
			Reboot:         true,
			TimeoutMinutes: 150, // Windows upgrade can take 60-90 min, + snapshot + reboot = 150 min total
			PrecheckDiskGB: 10,
		},
		Timeouts: TimeoutConfig{
			SignalScriptSeconds: 30,
			SignalFilesMinutes:  30,
			TargetOSMinutes:     20,
			PowerOffMinutes:     5,
		},
		Logging: LoggingConfig{
			Level: "info",
			File:  "osupgrader.log",
		},
		UI: UIConfig{
			Language: "en", // Default to English
			DarkMode: false,
		},
	}
}

func (cfg *AppConfig) applyTimeoutDefaults() {
	defaults := createDefaultConfig().Timeouts

	if cfg.Timeouts.SignalScriptSeconds == 0 {
		cfg.Timeouts.SignalScriptSeconds = defaults.SignalScriptSeconds
	}
	if cfg.Timeouts.SignalFilesMinutes == 0 {
		cfg.Timeouts.SignalFilesMinutes = defaults.SignalFilesMinutes
	}
	if cfg.Timeouts.TargetOSMinutes == 0 {
		cfg.Timeouts.TargetOSMinutes = defaults.TargetOSMinutes
	}
	if cfg.Timeouts.PowerOffMinutes == 0 {
		cfg.Timeouts.PowerOffMinutes = defaults.PowerOffMinutes
	}
}

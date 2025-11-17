package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// VCenterConfig innehåller vCenter-anslutningsinformation
type VCenterConfig struct {
	Host     string `json:"vcenter_url"`
	Username string `json:"username"`
	Password string `json:"-"` // Lagra aldrig lösenord i filen
	Session  string `json:"-"`
	Mode     string `json:"mode,omitempty"`     // "sspi" eller "password"
	Insecure bool   `json:"insecure,omitempty"` // tillåt osignerade cert
}

// DefaultsConfig för sektionen "defaults"
type DefaultsConfig struct {
	SnapshotNamePrefix   string `json:"snapshot_name_prefix"`
	IsoDatastorePath     string `json:"iso_datastore_path"`
	SkipMemoryInSnapshot bool   `json:"skip_memory_in_snapshot"`
	Glvk                 string `json:"glvk"`
	GuestUsername        string `json:"guest_username,omitempty"`
}

// UpgradeConfig för sektionen "upgrade"
type UpgradeConfig struct {
	Parallel       int  `json:"parallel"`
	Reboot         bool `json:"reboot"`
	TimeoutMinutes int  `json:"timeout_minutes"`
	PrecheckDiskGB int  `json:"precheck_disk_gb"`
}

// TimeoutConfig innehåller detaljerade timeout-inställningar
type TimeoutConfig struct {
	SignalScriptSeconds int `json:"signal_script_seconds"`
	SignalFilesMinutes  int `json:"signal_files_minutes"`
	TargetOSMinutes     int `json:"target_os_minutes"`
	PowerOffMinutes     int `json:"poweroff_minutes"`
}

// LoggingConfig för sektionen "logging"
type LoggingConfig struct {
	Level string `json:"level"`
	File  string `json:"file"`
}

// UIConfig för sektionen "ui"
type UIConfig struct {
	Language string `json:"language"`
	DarkMode bool   `json:"dark_mode"`
}

// AppConfig representerar konfigurationsfilens struktur
type AppConfig struct {
	VCenter  VCenterConfig  `json:"vcenter"`
	Defaults DefaultsConfig `json:"defaults"`
	Upgrade  UpgradeConfig  `json:"upgrade"`
	Timeouts TimeoutConfig  `json:"timeouts"`
	Logging  LoggingConfig  `json:"logging"`
	UI       UIConfig       `json:"ui"`
}

const configFileName = "conf.json"

// GetConfigPath returnerar sökvägen till konfigurationsfilen
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("kunde inte hitta home directory: %w", err)
	}
	return filepath.Join(homeDir, configFileName), nil
}

// Load läser in konfigurationen från fil
func Load() (*AppConfig, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Skapa defaultkonfiguration
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

// Save sparar konfigurationen till fil
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

// createDefaultConfig skapar en standardkonfiguration
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
			Glvk:                 "WX4NM-KYWYW-QJJR4-XV3QB-6VM33",
			GuestUsername:        "Administrator",
		},
		Upgrade: UpgradeConfig{
			Parallel:       10, // Antal parallella uppgraderingar
			Reboot:         true,
			TimeoutMinutes: 150, // Windows upgrade kan ta 60-90 min, + snapshot + reboot = 150 min total
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
			Language: "sv",
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

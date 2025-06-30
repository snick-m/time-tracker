package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	SpreadsheetID string `json:"spreadsheet_id"`
	SheetName     string `json:"sheet_name"`
	Hotkey        string `json:"hotkey"`
}

func LoadConfig() (*Config, error) {
	cfgPath, _ := os.UserConfigDir()
	cfgPath = filepath.Join(cfgPath, "time-tracker", "config.json")

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return &Config{
			Hotkey:        "ctrl+alt+q",
			SheetName:     "Sheet1",
			SpreadsheetID: "",
		}, nil
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	return &cfg, err
}

func SaveConfig(cfg *Config) error {
	cfgPath, _ := os.UserConfigDir()
	cfgPath = filepath.Join(cfgPath, "time-tracker")
	os.MkdirAll(cfgPath, 0755)

	data, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(filepath.Join(cfgPath, "config.json"), data, 0644)
}

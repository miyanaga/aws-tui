package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Settings struct {
	Favorites      []string `json:"favorites"`
	LocalDirectory string   `json:"local_directory,omitempty"`
}

func getSettingsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".aws-tui", "settings.json"), nil
}

func ensureSettingsDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(homeDir, ".aws-tui")
	return os.MkdirAll(dir, 0755)
}

func Load() (*Settings, error) {
	path, err := getSettingsPath()
	if err != nil {
		return &Settings{Favorites: []string{}}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Settings{Favorites: []string{}}, nil
		}
		return nil, err
	}

	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *Settings) Save() error {
	if err := ensureSettingsDir(); err != nil {
		return err
	}

	path, err := getSettingsPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (s *Settings) AddFavorite(view string) error {
	// Check if already exists
	for _, f := range s.Favorites {
		if f == view {
			return nil
		}
	}
	s.Favorites = append(s.Favorites, view)
	return s.Save()
}

func (s *Settings) RemoveFavorite(view string) error {
	for i, f := range s.Favorites {
		if f == view {
			s.Favorites = append(s.Favorites[:i], s.Favorites[i+1:]...)
			return s.Save()
		}
	}
	return nil
}

func (s *Settings) IsFavorite(view string) bool {
	for _, f := range s.Favorites {
		if f == view {
			return true
		}
	}
	return false
}

func (s *Settings) GetLocalDirectory() string {
	if s.LocalDirectory == "" {
		// Default to current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return os.Getenv("HOME")
		}
		return cwd
	}
	return s.LocalDirectory
}

func (s *Settings) SetLocalDirectory(dir string) error {
	// Validate directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return os.ErrInvalid
	}

	s.LocalDirectory = dir
	return s.Save()
}

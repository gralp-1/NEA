package main

import (
	"encoding/json"
	"os"
)

type Language int32

type FileFormat string

const (
	PNG  FileFormat = "png"
	JPG             = "jpg"
	TIFF            = "tiff"
	BMP             = "bmp"
)

type Config struct {
	Language   Language   `json:"Language"`
	FileFormat FileFormat `json:"FileFormat"`
}

func NewConfig() Config {
	return Config{Language: English, FileFormat: TIFF}
}

// TODO: make NEA config path configureable

func (c *Config) LoadConfigOrDefault() {
	// read ~/.neaconfig.json
	content, err := os.ReadFile("~/.neaconfig.json")
	state.Config = NewConfig()
	if err != nil {
		ErrorLogf("Error reading config: %v\nResorting to default config %v", err.Error(), state.Config)
		return
	}
	tempConfig := NewConfig()
	err = json.Unmarshal(content, &tempConfig)
	if err != nil {
		ErrorLogf("Error decoding config: %v\nResorting to default config %v", err.Error(), state.Config)
		return
	}
	state.Config = tempConfig
}

func (c *Config) SaveConfig() {
	content, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	_ = os.WriteFile("~/.imgconfig.json", content, 0666)
}

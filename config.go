package main

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

// Config is the configuration of the current device
type Config struct {
	XMLName         xml.Name         `xml:"config"`
	Router          *Router          `xml:"router"`
	Room            *Room            `xml:"room"`
	DeviceContainer *DeviceContainer `xml:"devices"`
	RouterContainer *RouterContainer `xml:"routers"`
}

// InitConfig returns either an unmarshalled Config struct
// or builds a Config that needs to be configured.
func InitConfig(path string) (*Config, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return BuildConfig()
	}
	return LoadConfig(path)
}

// LoadConfig returns a new Config struct given a path
func LoadConfig(path string) (*Config, error) {
	config := &Config{}
	configFile, configErr := ioutil.ReadFile(path)
	if configErr != nil {
		return nil, configErr
	}
	marshErr := xml.Unmarshal(configFile, config)
	if marshErr != nil {
		return nil, marshErr
	}
	return config, nil
}

// BuildConfig returns an unconfigured Config struct
func BuildConfig() (*Config, error) {
	router, routerErr := BuildRouter()
	if routerErr != nil {
		return nil, routerErr
	}
	room, roomErr := BuildRoom()
	if roomErr != nil {
		return nil, roomErr
	}
	config := &Config{
		Router: router,
		Room:   room,
	}
	return config, nil
}

// WriteConfig formats and exports the config struct to the
// file at the given location.
func (config *Config) WriteConfig(path string) error {
	configData, configErr := xml.MarshalIndent(config, "", "    ")
	if configErr != nil {
		return configErr
	}
	writeErr := ioutil.WriteFile(path, configData, 0644)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

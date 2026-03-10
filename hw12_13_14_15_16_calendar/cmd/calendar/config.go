package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger  LoggerConf `yaml:"logger"`
	Storage Storage    `yaml:"storage"`
	Server  Server     `yaml:"server"`
}

type LoggerConf struct {
	Level string `yaml:"level"`
}

type Storage struct {
	Mode string `yaml:"mode"`
	Dsn  string `yaml:"dsn"`
}

type Server struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func NewConfig() Config {
	return Config{}
}

func (c *Config) readConfig(configFile string) error {
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("error read file: %w", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return fmt.Errorf("error unmarshal yaml: %w", err)
	}
	return nil
}

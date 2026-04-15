package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Logger  LoggerConf `yaml:"logger"`
	Storage Storage    `yaml:"storage"`
	Kafka   Kafka      `yaml:"kafka"`
}

type LoggerConf struct {
	Level    string `yaml:"level"`
	Filename string `yaml:"filename"`
}

type Storage struct {
	Mode string `yaml:"mode"`
	Dsn  string `yaml:"dsn"`
}

type Kafka struct {
	Brokers []string      `yaml:"brokers" default:"localhost:9092"`
	GroupID string        `yaml:"group-id" default:"storer-group"`
	Timeout time.Duration `yaml:"timeout" default:"10s"`
	Topics  struct {
		Notifications string `yaml:"topic_notifications" default:"notifications"`
	} `yaml:"topics"`
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

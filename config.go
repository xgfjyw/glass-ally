package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Key  string `yaml:"key"`
	Cx   string `yaml:"cx"`
	Dsn  string `yaml:"db_dsn"`
	Addr string `yaml:"listen"`
}

func initConfig() Config {
	cfgFile, err := os.Open("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer cfgFile.Close()

	config := Config{}
	cfg := yaml.NewDecoder(cfgFile)
	err = cfg.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

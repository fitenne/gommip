package main

import (
	"flag"

	"go.uber.org/zap"
)

func ParseConfig() *Config {
	configFilePath := flag.String("config", "config.yml", "path to config file")
	flag.Parse()

	return configFrom(*configFilePath)
}

func main() {
	config := ParseConfig()
	initLogger(config)
	if err := prepareDb(config); err != nil {
		zap.S().Fatalln(err)
	}
	if err := serve(config); err != nil {
		zap.S().Fatalln(err)
	}
}

func serve(config *Config) error {
	return NewServer(config).Run(config.Listen)
}

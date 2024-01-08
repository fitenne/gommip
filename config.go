package main

import (
	"log"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

type Download struct {
	Dir       string
	ReloadGap time.Duration `yaml:"reload_gap"`
	AsnUrl    string        `yaml:"asn_url"`
	CityUrl   string        `yaml:"city_url"`
}

type Log struct {
	Level  string
	Access string
	Output []string
}

type Config struct {
	*Download
	*Log
	Listen       string
	RealIpHeader []string `yaml:"real_ip_header"`
}

func configFrom(configFilePath string) *Config {
	cf, err := os.Open(configFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer cf.Close()

	var conf Config
	if err = yaml.NewDecoder(cf).Decode(&conf); err != nil {
		log.Fatalln(err)
		return nil
	}

	return &conf
}

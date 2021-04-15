package config

import (
	"flag"
	"gopkg.in/yaml.v2"
	"os"
)

type Configuration struct {
	DbDir      string `yaml:"db_dir"`
	ServerAddr string `yaml:"server_addr"`
	Debug      bool   `yaml:"debug"`
}

var Config = Configuration{
	DbDir:      "/var/lib/v2stat",
	ServerAddr: "127.0.0.1:10085",
	Debug:      false,
}

var confPath = flag.String("conf", "/etc/v2stat.yaml", "Path to the config file")

func LoadConf() error {
	flag.Parse()

	// Save default config if not exist
	if _, err := os.Stat(*confPath); os.IsNotExist(err) {
		return SaveConf()
	}

	f, err := os.Open(*confPath)
	if err != nil {
		return err
	}
	defer f.Close()
	err = yaml.NewDecoder(f).Decode(&Config)
	return err
}

func SaveConf() error {
	f, err := os.Create(*confPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewEncoder(f).Encode(Config)
}

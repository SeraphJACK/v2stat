package config

import (
	"flag"
	"gopkg.in/yaml.v2"
	"os"
)

type Configuration struct {
	DbDir          string `yaml:"db_dir"`
	ServerAddr     string `yaml:"server_addr"`
	DaysToKeep     int    `yaml:"days_to_keep"`
	MonthsToKeep   int    `yaml:"months_to_keep"`
	YearsToKeep    int    `yaml:"years_to_keep"`
	ResetOnStart   bool   `yaml:"reset_on_start"`
	Debug          bool   `yaml:"debug"`
	MemProfilePath string `yaml:"mem_profile_path"`
}

var Config = Configuration{
	DbDir:          "/var/lib/v2stat",
	ServerAddr:     "127.0.0.1:10085",
	DaysToKeep:     7,
	MonthsToKeep:   2,
	YearsToKeep:    3,
	ResetOnStart:   false,
	Debug:          false,
	MemProfilePath: "/var/lib/v2stat/mem.prof",
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

package app

import (
	"github.com/theherk/viper"
	"log"
	"path/filepath"
	"strings"
)

// Initialize config.toml config. Fatal if not exist.
func InitConfig(f string) {
	name, dir, ext := filepath.Base(f), filepath.Dir(f), filepath.Ext(f)
	name = strings.TrimSuffix(name, ext)

	viper.SetConfigName(name)
	viper.AddConfigPath(dir)
	err := viper.ReadInConfig()

	if err != nil { // Handle errors reading the config file
		log.Fatalf("Fatal error config file: %s \n", err)
	}
}

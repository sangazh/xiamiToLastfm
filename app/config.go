package app

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
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

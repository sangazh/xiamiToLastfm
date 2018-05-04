package app

import (
	"github.com/theherk/viper"
	"log"
)

// Initialize config.toml config. Fatal if not exist.
func InitConfig() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	err := viper.ReadInConfig()   // Find and read the config file

	if err != nil { // Handle errors reading the config file
		log.Fatalf("Fatal error config file: %s \n", err)
	}
}
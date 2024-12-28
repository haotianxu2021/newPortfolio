package util

import (
	"fmt"

	"github.com/spf13/viper" // Replaced godotenv with viper
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

// LoadConfig reads configuration from environment variables or file
func LoadConfig() (config Config, err error) {
	v := viper.New()
	// v.AutomaticExpansion() // Enable automatic expansion of environment variables

	// Set the path for viper to look for the config file
	v.SetConfigFile("./.env")
	v.SetConfigName("")    // Name of config file (without extension)
	v.SetConfigType("env") // Config file type
	v.AddConfigPath(".")   // Path to look for the config file (current directory)

	// Load the config file
	if err := v.ReadInConfig(); err != nil {
		// It's okay if .env doesn't exist in production
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	config.DBDriver = v.GetString("DB_DRIVER")
	if config.DBDriver == "" {
		config.DBDriver = "postgres" // default value
	}

	config.DBSource = v.GetString("DB_SOURCE")
	if config.DBSource == "" {
		return config, fmt.Errorf("DB_SOURCE environment variable is required")
	}

	config.ServerAddress = v.GetString("SERVER_ADDRESS")
	if config.ServerAddress == "" {
		config.ServerAddress = ":8080" // default value
	}

	return config, nil
}

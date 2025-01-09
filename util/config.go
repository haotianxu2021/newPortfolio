package util

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

// Config stores all configuration of the application.
// The values are read from environment variables.
type Config struct {
	DBDriver            string        `mapstructure:"DB_DRIVER"`
	DBSource            string        `mapstructure:"DB_SOURCE"`
	ServerAddress       string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey   string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
}

// loadEnvFile reads and parses the .env file if it exists.
// It loads environment variables into the process environment but does not return a Config struct.
// Use LoadConfig to get the final configuration with environment variables applied.
func loadEnvFile(filename string) error {
	// Try current directory first
	file, err := os.Open(filename)
	if err == nil {
		log.Printf("loaded .env file from current directory")
		defer file.Close()
		return parseEnvFile(file)
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("error opening .env file: %w", err)
	}

	// If not found, try parent directories
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working directory: %w", err)
	}

	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			return nil
		}
		dir = parent

		envPath := filepath.Join(dir, filename)
		file, err := os.Open(envPath)
		if err == nil {
			defer file.Close()
			return parseEnvFile(file)
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("error opening .env file at %s: %w", envPath, err)
		}
	}
}

// parseEnvFile parses an open .env file
func parseEnvFile(file *os.File) error {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 2 {
			continue // Skip malformed lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove any quotes
		value = strings.Trim(value, `"'`)

		// Clean the key - only allow alphanumeric characters and underscore
		key = strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
				return r
			}
			return -1
		}, key)

		// Validate key is not empty after cleaning
		if key == "" {
			log.Printf("Warning: skipping invalid environment key")
			continue
		}

		// Clean the value - remove any control characters
		value = strings.Map(func(r rune) rune {
			if unicode.IsControl(r) {
				return -1
			}
			return r
		}, value)

		// log.Printf("Setting environment variable: key='%s', value='%s'", key, value)

		// Set the environment variable
		if err := os.Setenv(key, value); err != nil {
			log.Printf("Warning: failed to set environment variable %s: %v", key, err)
			continue // Skip this variable but continue processing others
		}
	}

	return scanner.Err()
}

// LoadConfig reads configuration from environment variables or file
func LoadConfig() (config Config, err error) {
	// Try loading .env file but don't fail if it doesn't exist
	if err := loadEnvFile(".env"); err != nil && !os.IsNotExist(err) {
		// Log the error but continue
		fmt.Printf("Warning: failed to load .env file: %v\n", err)
	}

	// Load all config values from environment variables
	config.DBDriver = os.Getenv("DB_DRIVER")
	config.DBSource = os.Getenv("DB_SOURCE")
	config.ServerAddress = os.Getenv("SERVER_ADDRESS")
	config.TokenSymmetricKey = os.Getenv("TOKEN_SYMMETRIC_KEY")

	// Parse duration if set
	if durationStr := os.Getenv("ACCESS_TOKEN_DURATION"); durationStr != "" {
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return config, fmt.Errorf("invalid ACCESS_TOKEN_DURATION format: %w", err)
		}
		config.AccessTokenDuration = duration
	}
	// log.Printf("%s, %s, %s, %s, %s", config.DBDriver, config.DBSource, config.ServerAddress, config.TokenSymmetricKey, config.AccessTokenDuration)

	// Apply defaults and validate
	if config.DBDriver == "" {
		config.DBDriver = "postgres" // default value
	}

	if config.AccessTokenDuration == 0 {
		config.AccessTokenDuration = 15 * time.Minute // default value
	}

	if config.DBSource == "" {
		return config, fmt.Errorf("DB_SOURCE environment variable is required")
	}

	if config.ServerAddress == "" {
		config.ServerAddress = ":8080" // default value
	}

	return config, nil
}

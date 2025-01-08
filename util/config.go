package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config stores all configuration of the application.
// The values are read from environment variables.
type Config struct {
	DBDriver      string
	DBSource      string
	ServerAddress string
}

// loadEnvFile reads and parses the .env file if it exists
func loadEnvFile(filename string) error {
	// Try to find .env in current directory and parent directories
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working directory: %w", err)
	}

	// Keep going up until we find .env or hit the root
	for {
		envPath := dir + string(os.PathSeparator) + filename
		file, err := os.Open(envPath)
		if err == nil {
			defer file.Close()
			return parseEnvFile(file)
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("error opening .env file: %w", err)
		}

		// Go up one directory
		lastSep := strings.LastIndex(dir, string(os.PathSeparator))
		if lastSep == -1 {
			// We've hit the root
			return nil
		}
		dir = dir[:lastSep]
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
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		if os.Getenv(key) == "" { // Only set if not already set in environment
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// LoadConfig reads configuration from environment variables or file
func LoadConfig() (config Config, err error) {
	// Try to load .env file first
	if err := loadEnvFile(".env"); err != nil {
		return config, err
	}

	// Set DBDriver with default value
	config.DBDriver = os.Getenv("DB_DRIVER")
	if config.DBDriver == "" {
		config.DBDriver = "postgres" // default value
	}

	// Set DBSource (required)
	config.DBSource = os.Getenv("DB_SOURCE")
	if config.DBSource == "" {
		return config, fmt.Errorf("DB_SOURCE environment variable is required")
	}

	// Set ServerAddress with default value
	config.ServerAddress = os.Getenv("SERVER_ADDRESS")
	if config.ServerAddress == "" {
		config.ServerAddress = ":8080" // default value
	}

	return config, nil
}

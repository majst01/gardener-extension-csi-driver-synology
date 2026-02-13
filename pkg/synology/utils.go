package synology

import (
	"fmt"
)

// Credentials holds Synology credentials
type Credentials struct {
	Host     string
	Port     int
	UseSSL   bool
	Username string
	Password string
}

// ValidateCredentials validates Synology credentials
func ValidateCredentials(creds *Credentials) error {
	if creds.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if creds.Port <= 0 || creds.Port > 65535 {
		return fmt.Errorf("invalid port: %d", creds.Port)
	}
	if creds.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if creds.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	return nil
}

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	validYAML = `
---
ddns:
  log-level: debug
  check-every-seconds: 300
  process-timeout-seconds: 20
  aws:
    - account: example
      region: us-east-1
      access-key: "1234567890"
      secret-key: "1234567890"
      domains:
        - name: "example.net"
          zone-id: Z0123456789ABCDEF
          records:
            - fqdn: vpn.example.net.
              record-type: A
              record-ttl: 3600
            - fqdn: vpn.example.net.
              record-type: AAAA
              record-ttl: 3600
            - fqdn: test-a.example.net.
              record-type: A
              record-ttl: 3600
acme:
  log-level: debug
  aws:
    - account: example
      region: us-east-1
      access-key: "1234567890"
      secret-key: "1234567890"
      domains:
        - name: "example.net"
          zone-id: Z0123456789ABCDEF
        - name: "vpn.example.net"
          zone-id: Z0123456789ABCDEF
  store:
    type: "sqlite"
    sqlite:
      file: "/tmp/acme.db"
`

	invalidYAML = `
ddns:
log-level: "debug"
check-every-seconds: 300
process-timeout-seconds: 20
  aws:
    - region: "us-east-1"
      zones:
        - id: "Z0123456789ABCDEF"
          records:
          - fqdn: "vpn.example.net."
            record-type: A
            record-ttl: 3600	
	`
)

// withConfig writes content (if non-nil) as config.yaml into a temporary
// directory and points the package search paths to it for the duration
// of the test.
func withConfig(t *testing.T, content string) {
	t.Helper()

	dir := t.TempDir()
	if content != "" {
		file := filepath.Join(dir, configName+"."+configFileType)
		if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
			t.Fatalf("writing config file: %v", err)
		}
	}

	original := paths
	paths = []string{dir}
	t.Cleanup(func() { paths = original })
}

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		yaml          string
		expectedError error
	}{
		{
			name:          "valid yaml",
			yaml:          validYAML,
			expectedError: nil,
		},
		{
			name:          "invalid yaml",
			yaml:          invalidYAML,
			expectedError: ErrReadConfig,
		},
	}

	for _, tt := range tests {
		expectedError := tt.expectedError

		t.Run(tt.name, func(t *testing.T) {
			withConfig(t, tt.yaml)

			_, err := New()

			if expectedError != nil {
				assert.ErrorIs(t, err, expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name          string
		yaml          string
		key           string
		value         any
		expectedValue any
		expectedError error
	}{
		{
			name: "valid-yaml-key-acme-found",
			yaml: validYAML,
			key:  "acme",
			value: &struct {
				Store struct {
					Type   string `mapstructure:"type"`
					Sqlite struct {
						File string `mapstructure:"file"`
					} `mapstructure:"sqlite"`
				} `mapstructure:"store"`
			}{},
			expectedValue: &struct {
				Store struct {
					Type   string `mapstructure:"type"`
					Sqlite struct {
						File string `mapstructure:"file"`
					} `mapstructure:"sqlite"`
				} `mapstructure:"store"`
			}{
				Store: struct {
					Type   string `mapstructure:"type"`
					Sqlite struct {
						File string `mapstructure:"file"`
					} `mapstructure:"sqlite"`
				}{
					Type: "sqlite",
					Sqlite: struct {
						File string `mapstructure:"file"`
					}{
						File: "/tmp/acme.db",
					},
				},
			},
		},
		{
			name: "valid-yaml-key-ddns-found",
			yaml: validYAML,
			key:  "ddns",
			value: &struct {
				LogLevel          string `mapstructure:"log-level"`
				CheckEverySeconds int    `mapstructure:"check-every-seconds"`
				ProcessTimeout    int    `mapstructure:"process-timeout-seconds"`
			}{},
			expectedValue: &struct {
				LogLevel          string `mapstructure:"log-level"`
				CheckEverySeconds int    `mapstructure:"check-every-seconds"`
				ProcessTimeout    int    `mapstructure:"process-timeout-seconds"`
			}{
				LogLevel:          "debug",
				CheckEverySeconds: 300,
				ProcessTimeout:    20,
			},
			expectedError: nil,
		},
		{
			name: "valid yaml key not found",
			yaml: validYAML,
			key:  "ddns-not-found",
			value: &struct {
				LogLevel          string `mapstructure:"log-level"`
				CheckEverySeconds int    `mapstructure:"check-every-seconds"`
				ProcessTimeout    int    `mapstructure:"process-timeout-seconds"`
				KeyNotFound       string `mapstructure:"key-not-found"`
			}{},
			expectedValue: &struct {
				LogLevel          string `mapstructure:"log-level"`
				CheckEverySeconds int    `mapstructure:"check-every-seconds"`
				ProcessTimeout    int    `mapstructure:"process-timeout-seconds"`
				KeyNotFound       string `mapstructure:"key-not-found"`
			}{
				LogLevel:          "",
				CheckEverySeconds: 0,
				ProcessTimeout:    0,
				KeyNotFound:       "",
			},
			expectedError: ErrKeyNotFound,
		},
		{
			name:          "valid yaml unmarshal error",
			yaml:          validYAML,
			key:           "ddns",
			value:         new(string),
			expectedValue: new(string),
			expectedError: ErrUnmarshal,
		},
	}

	for _, tt := range tests {
		expectedError := tt.expectedError
		key := tt.key
		value := tt.value
		expectedValue := tt.expectedValue

		t.Run(tt.name, func(t *testing.T) {
			withConfig(t, tt.yaml)

			cnf, err := New()
			assert.NoError(t, err)

			err = cnf.Get(key, value)

			assert.Equal(t, expectedValue, value)
			if expectedError != nil {
				assert.ErrorIs(t, err, expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

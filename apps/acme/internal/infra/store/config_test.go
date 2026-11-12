package store

import (
	"testing"

	"github.com/go-viper/mapstructure/v2"
	"github.com/stretchr/testify/assert"
)

type configuratorMock struct {
	value       map[string]any
	err         error
	receivedKey string
}

func (cm *configuratorMock) Get(key string, value any) error {
	cm.receivedKey = key

	if cm.err != nil {
		return cm.err
	}

	return mapstructure.Decode(cm.value, value)
}

// func readConfig(cfg Configurator) (*storeConfig, error) {
func TestReadConfig(t *testing.T) {
	testCases := []struct {
		name            string
		expectedKey     string
		storeConfigMock *configuratorMock
		expectedResult  *storeConfig
		expectedError   error
	}{
		{
			name:          "normal-case",
			expectedKey:   rootConfigKey,
			expectedError: nil,
			expectedResult: &storeConfig{
				Type: "sqlite",
				SQLite: &sqliteConfig{
					File: "/tmp/db.sqlite",
				},
			},
			storeConfigMock: &configuratorMock{
				value: map[string]any{
					"type": "sqlite",
					"sqlite": map[string]any{
						"file": "/tmp/db.sqlite",
					},
				},
			},
		},
		{
			name:           "no-sqlite-info-case",
			expectedKey:    rootConfigKey,
			expectedError:  ErrInvalidConfigurationValue,
			expectedResult: nil,
			storeConfigMock: &configuratorMock{
				value: map[string]any{
					"type":   "sqlite",
					"sqlite": map[string]any{},
				},
			},
		},
	}

	for _, tt := range testCases {
		mock := tt.storeConfigMock
		expectedError := tt.expectedError
		expectedResult := tt.expectedResult
		expectedKey := tt.expectedKey

		t.Run(tt.name, func(t *testing.T) {
			//t.Parallel()

			result, err := readConfig(mock)

			assert.Equal(t, expectedKey, mock.receivedKey)
			assert.Equal(t, expectedResult, result)

			if expectedError != nil {
				assert.ErrorIs(t, err, expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name          string
		cfg           *storeConfig
		expectedError error
	}{
		{
			name: "valid-config",
			cfg: &storeConfig{
				Type:   "sqlite",
				SQLite: &sqliteConfig{File: "/tmp/db.sqlite"},
			},
			expectedError: nil,
		},
		{
			name:          "nil-config",
			cfg:           nil,
			expectedError: ErrInvalidConfiguration,
		},
		{
			name: "missing-sqlite-file",
			cfg: &storeConfig{
				Type:   "sqlite",
				SQLite: &sqliteConfig{},
			},
			expectedError: ErrInvalidConfigurationValue,
		},
		{
			name: "unsupported-type",
			cfg: &storeConfig{
				Type:   "mysql",
				SQLite: &sqliteConfig{File: "/tmp/db.sqlite"},
			},
			expectedError: ErrInvalidConfigurationValue,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateConfig(tt.cfg)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	testCases := []struct {
		name                   string
		mockedGodotenvLoad     func(filenames ...string) (err error)
		mockedEnvconfigProcess func(prefix string, spec interface{}) error
		expectedError          error
	}{
		{
			name: "happy path",
			mockedGodotenvLoad: func(filenames ...string) (err error) {
				return nil
			},
			mockedEnvconfigProcess: func(prefix string, spec interface{}) error {
				return nil
			},
		},
		{
			name: "error loading env vars",
			mockedGodotenvLoad: func(filenames ...string) (err error) {
				return errors.New("random error")
			},
			expectedError: errors.New("loading env vars from .env file: random error"),
		},
		{
			name: "error processing env vars",
			mockedGodotenvLoad: func(filenames ...string) (err error) {
				return nil
			},
			mockedEnvconfigProcess: func(prefix string, spec interface{}) error {
				return errors.New("random error")
			},
			expectedError: errors.New("processing env vars: random error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			godotenvLoad = tc.mockedGodotenvLoad
			envconfigProcess = tc.mockedEnvconfigProcess
			config, err := Read()
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf("expected no error, got %v", err)
				}
				require.Nil(t, config)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf("expected error, got nil")
				}
				require.NotNil(t, config)
			}
		})
	}
}

func TestReadFromEnvFile(t *testing.T) {
	testCases := []struct {
		name                   string
		mockedGodotenvLoad     func(filenames ...string) (err error)
		mockedEnvconfigProcess func(prefix string, spec interface{}) error
		expectedError          error
	}{
		{
			name: "happy path",
			mockedGodotenvLoad: func(filenames ...string) (err error) {
				return nil
			},
			mockedEnvconfigProcess: func(prefix string, spec interface{}) error {
				return nil
			},
		},
		{
			name: "error loading env vars",
			mockedGodotenvLoad: func(filenames ...string) (err error) {
				return errors.New("random error")
			},
			expectedError: errors.New("loading env vars from path/to/.env: random error"),
		},
		{
			name: "error processing env vars",
			mockedGodotenvLoad: func(filenames ...string) (err error) {
				return nil
			},
			mockedEnvconfigProcess: func(prefix string, spec interface{}) error {
				return errors.New("random error")
			},
			expectedError: errors.New("processing env vars: random error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			godotenvLoad = tc.mockedGodotenvLoad
			envconfigProcess = tc.mockedEnvconfigProcess
			config, err := ReadFromEnvFile("path/to/.env")
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf("expected no error, got %v", err)
				}
				require.Nil(t, config)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf("expected error, got nil")
				}
				require.NotNil(t, config)
			}
		})
	}
}

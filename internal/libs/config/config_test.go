package config

import (
	"os"
	"reflect"
	"testing"
)

func TestGetBindings(t *testing.T) {
	tests := []struct {
		name     string
		tag      reflect.StructTag
		expected []string
	}{
		{
			name:     "No bindings",
			tag:      `binding:""`,
			expected: []string{},
		},
		{
			name:     "Single binding",
			tag:      `binding:"required"`,
			expected: []string{"required"},
		},
		{
			name:     "Multiple bindings",
			tag:      `binding:"required,optional"`,
			expected: []string{"required", "optional"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBindings(tt.tag)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
func TestIsRequired(t *testing.T) {
	tests := []struct {
		name     string
		tag      reflect.StructTag
		expected bool
	}{
		{
			name:     "No bindings",
			tag:      `binding:""`,
			expected: false,
		},
		{
			name:     "Required binding",
			tag:      `binding:"required"`,
			expected: true,
		},
		{
			name:     "Optional binding",
			tag:      `binding:"optional"`,
			expected: false,
		},
		{
			name:     "Multiple bindings with required",
			tag:      `binding:"required,optional"`,
			expected: true,
		},
		{
			name:     "Multiple bindings without required",
			tag:      `binding:"optional,another"`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRequired(tt.tag)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
func TestGetValueFromString(t *testing.T) {
	tests := []struct {
		name         string
		val          string
		expectedType reflect.Type
		expected     any
		expectErr    bool
	}{
		{
			name:         "String type",
			val:          "test",
			expectedType: reflect.TypeOf(""),
			expected:     "test",
			expectErr:    false,
		},
		{
			name:         "Int type",
			val:          "123",
			expectedType: reflect.TypeOf(0),
			expected:     123,
			expectErr:    false,
		},
		{
			name:         "Bool type true",
			val:          "true",
			expectedType: reflect.TypeOf(true),
			expected:     true,
			expectErr:    false,
		},
		{
			name:         "Bool type false",
			val:          "false",
			expectedType: reflect.TypeOf(false),
			expected:     false,
			expectErr:    false,
		},
		{
			name:         "Unsupported type",
			val:          "unsupported",
			expectedType: reflect.TypeOf([]string{}),
			expected:     nil,
			expectErr:    true,
		},
		{
			name:         "Invalid int",
			val:          "abc",
			expectedType: reflect.TypeOf(0),
			expected:     nil,
			expectErr:    true,
		},
		{
			name:         "Invalid bool",
			val:          "notabool",
			expectedType: reflect.TypeOf(true),
			expected:     nil,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getValueFromString(tt.val, tt.expectedType)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
			if !tt.expectErr && !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
func TestGetFromDefault(t *testing.T) {
	tests := []struct {
		name         string
		tag          reflect.StructTag
		expectedType reflect.Type
		expected     any
		expectErr    bool
	}{
		{
			name:         "Default string value",
			tag:          `default:"defaultValue"`,
			expectedType: reflect.TypeOf(""),
			expected:     "defaultValue",
			expectErr:    false,
		},
		{
			name:         "Default int value",
			tag:          `default:"42"`,
			expectedType: reflect.TypeOf(0),
			expected:     42,
			expectErr:    false,
		},
		{
			name:         "Default bool value true",
			tag:          `default:"true"`,
			expectedType: reflect.TypeOf(true),
			expected:     true,
			expectErr:    false,
		},
		{
			name:         "Default bool value false",
			tag:          `default:"false"`,
			expectedType: reflect.TypeOf(false),
			expected:     false,
			expectErr:    false,
		},
		{
			name:         "Unsupported type",
			tag:          `default:"unsupported"`,
			expectedType: reflect.TypeOf([]string{}),
			expected:     nil,
			expectErr:    true,
		},
		{
			name:         "Invalid int value",
			tag:          `default:"notanint"`,
			expectedType: reflect.TypeOf(0),
			expected:     nil,
			expectErr:    true,
		},
		{
			name:         "Invalid bool value",
			tag:          `default:"notabool"`,
			expectedType: reflect.TypeOf(true),
			expected:     nil,
			expectErr:    true,
		},
		{
			name:         "No default value",
			tag:          `default:""`,
			expectedType: reflect.TypeOf(""),
			expected:     nil,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getFromDefault(tt.tag, tt.expectedType)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
			if !tt.expectErr && !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
func TestGetFromEnv(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		tag          reflect.StructTag
		expectedType reflect.Type
		expected     any
		expectErr    bool
	}{
		{
			name:         "String from env",
			envKey:       "TEST_STRING",
			envValue:     "testValue",
			tag:          `env:"TEST_STRING"`,
			expectedType: reflect.TypeOf(""),
			expected:     "testValue",
			expectErr:    false,
		},
		{
			name:         "Int from env",
			envKey:       "TEST_INT",
			envValue:     "123",
			tag:          `env:"TEST_INT"`,
			expectedType: reflect.TypeOf(0),
			expected:     123,
			expectErr:    false,
		},
		{
			name:         "Bool true from env",
			envKey:       "TEST_BOOL_TRUE",
			envValue:     "true",
			tag:          `env:"TEST_BOOL_TRUE"`,
			expectedType: reflect.TypeOf(true),
			expected:     true,
			expectErr:    false,
		},
		{
			name:         "Bool false from env",
			envKey:       "TEST_BOOL_FALSE",
			envValue:     "false",
			tag:          `env:"TEST_BOOL_FALSE"`,
			expectedType: reflect.TypeOf(false),
			expected:     false,
			expectErr:    false,
		},
		{
			name:         "Env key not found",
			envKey:       "",
			envValue:     "",
			tag:          `env:"NON_EXISTENT_KEY"`,
			expectedType: reflect.TypeOf(""),
			expected:     nil,
			expectErr:    true,
		},
		{
			name:         "Invalid int from env",
			envKey:       "INVALID_INT",
			envValue:     "notanint",
			tag:          `env:"INVALID_INT"`,
			expectedType: reflect.TypeOf(0),
			expected:     nil,
			expectErr:    true,
		},
		{
			name:         "Invalid bool from env",
			envKey:       "INVALID_BOOL",
			envValue:     "notabool",
			tag:          `env:"INVALID_BOOL"`,
			expectedType: reflect.TypeOf(true),
			expected:     nil,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result, err := getFromEnv(tt.tag, tt.expectedType)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
			if !tt.expectErr && !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
func TestLoad(t *testing.T) {
	type Config struct {
		StringField string `env:"TEST_STRING" binding:"required"`
		IntField    int    `env:"TEST_INT" default:"42"`
		BoolField   bool   `env:"TEST_BOOL" default:"true"`
	}

	tests := []struct {
		name      string
		envVars   map[string]string
		expected  Config
		expectErr bool
	}{
		{
			name: "All fields from env",
			envVars: map[string]string{
				"TEST_STRING": "envString",
				"TEST_INT":    "123",
				"TEST_BOOL":   "false",
			},
			expected: Config{
				StringField: "envString",
				IntField:    123,
				BoolField:   false,
			},
			expectErr: false,
		},
		{
			name: "Default values",
			envVars: map[string]string{
				"TEST_STRING": "test",
			},
			expected: Config{
				StringField: "test",
				IntField:    42,
				BoolField:   true,
			},
			expectErr: false,
		},
		{
			name: "Missing required field",
			envVars: map[string]string{
				"TEST_INT":  "123",
				"TEST_BOOL": "false",
			},
			expected:  Config{},
			expectErr: true,
		},
		{
			name: "Invalid int value",
			envVars: map[string]string{
				"TEST_STRING": "envString",
				"TEST_INT":    "notanint",
				"TEST_BOOL":   "false",
			},
			expected:  Config{},
			expectErr: true,
		},
		{
			name: "Invalid bool value",
			envVars: map[string]string{
				"TEST_STRING": "envString",
				"TEST_INT":    "123",
				"TEST_BOOL":   "notabool",
			},
			expected:  Config{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				if v != "" {
					t.Setenv(k, v)
				}
			}

			var cfg Config
			err := Load(&cfg)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
			if !tt.expectErr && !reflect.DeepEqual(cfg, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, cfg)
			}
		})
	}
}

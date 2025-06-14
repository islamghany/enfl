package enfl

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"
)

// enfl is library for loading environment variables from .env file and command line arguments

// Loader handles configuration loading from multiple sources
type Loader struct {
	envPrefix   string
	flagSet     *flag.FlagSet // flag set for command line arguments
	failOnError bool
	envFiles    []string
	autoLoadEnv bool
}

type Option func(*Loader)

// WithFlagSet sets a custom flag set
func WithFlagSet(flagSet *flag.FlagSet) Option {
	return func(l *Loader) {
		l.flagSet = flagSet
	}
}

// WithEnvPrefix sets a global environment variable prefix
func WithEnvPrefix(prefix string) Option {
	return func(l *Loader) {
		l.envPrefix = prefix
	}
}

// WithFailOnError makes the loader fail fast on errors
func WithFailOnError(failOnError bool) Option {
	return func(l *Loader) {
		l.failOnError = failOnError
	}
}

// WithEnvFiles specifies .env files to load
func WithEnvFiles(files ...string) Option {
	return func(l *Loader) {
		l.envFiles = files
	}
}

// WithAutoLoadEnv automatically looks for common .env files
func WithAutoLoadEnv(autoLoadEnv bool) Option {
	return func(l *Loader) {
		l.autoLoadEnv = autoLoadEnv
	}
}

// NewLoader creates a new loader with default options
func NewLoader(opts ...Option) *Loader {
	l := &Loader{
		flagSet:     flag.CommandLine,
		failOnError: true,
		autoLoadEnv: true,
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// Load loads environment variables from .env files and command line arguments into the given config struct
func (l *Loader) Load(config interface{}) error {
	// check if config is a pointer
	v := reflect.ValueOf(config)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	// Load .env files (lowest priority after defaults)
	if err := l.loadEnvFiles(); err != nil {
		if l.failOnError {
			return fmt.Errorf("failed to load .env files: %w", err)
		}
		fmt.Fprintf(os.Stderr, "config warning: failed to load .env files: %v\n", err)
	}
	return l.processStruct(v.Elem(), "")
}

// processStruct processes a struct and its fields
func (l *Loader) processStruct(v reflect.Value, prefix string) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Handle nested structs
		if field.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Time{}) {
			nestedPrefix := l.getNestedPrefix(fieldType, prefix)
			if err := l.processStruct(field, nestedPrefix); err != nil {
				return err
			}
			continue
		}

		if err := l.processField(field, fieldType, prefix); err != nil {
			if l.failOnError {
				return err
			}
			// Log error but continue
			fmt.Fprintf(os.Stderr, "config warning: %v\n", err)
		}
	}
	return nil
}

// getNestedPrefix gets the prefix for nested structs
func (l *Loader) getNestedPrefix(field reflect.StructField, currentPrefix string) string {
	if prefixTag := field.Tag.Get("prefix"); prefixTag != "" {
		return currentPrefix + prefixTag
	}
	return currentPrefix + toSnakeCase(field.Name) + "_"
}

// loadEnvFiles loads environment variables from .env files
func (l *Loader) loadEnvFiles() error {
	filesToLoad := l.envFiles
	if l.autoLoadEnv {
		commonEnvFiles := []string{
			".env",
			".env.local",
			".env.development",
			".env.production",
		}
		// Only include files that exist
		for _, file := range commonEnvFiles {
			if _, err := os.Stat(file); err == nil {
				filesToLoad = append(filesToLoad, file)
			}
		}
	}

	// Load each file
	for _, file := range filesToLoad {
		if err := l.loadEnvFile(file); err != nil {
			return fmt.Errorf("failed to load %s: %w", file, err)
		}
	}

	return nil
}

// loadEnvFile loads a single .env file
func (l *Loader) loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format at line %d: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Skip if key is empty
		if key == "" || value == "" {
			continue
		}

		// Handle Quated Values
		value = l.unquoteValue(value)
		// Only set if not already set (environment variables take precedence)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// unquoteValue removes quotes from values and handles escape sequences
func (l *Loader) unquoteValue(value string) string {
	// Handle double quotes
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
		// Handle escape sequences
		value = strings.ReplaceAll(value, `\"`, `"`)
		value = strings.ReplaceAll(value, `\\`, `\`)
		value = strings.ReplaceAll(value, `\n`, "\n")
		value = strings.ReplaceAll(value, `\r`, "\r")
		value = strings.ReplaceAll(value, `\t`, "\t")
		return value
	}

	// Handle single quotes (no escape sequences)
	if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
		return value[1 : len(value)-1]
	}

	return value
}

// Convenience functions
func Load(config interface{}) error {
	return NewLoader().Load(config)
}

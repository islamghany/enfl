package enfl

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
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
	// Register flags first
	if err := l.registerFlags(v.Elem(), ""); err != nil {
		return fmt.Errorf("failed to register flags: %w", err)
	}

	// Parse command line flags if using CommandLine and not already parsed
	if l.flagSet == flag.CommandLine && !flag.Parsed() {
		flag.Parse()
	}

	return l.processStruct(v.Elem(), "")
}

// registerFlags registers all flags with the flag set
func (l *Loader) registerFlags(v reflect.Value, prefix string) error {
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
			if err := l.registerFlags(field, nestedPrefix); err != nil {
				return err
			}
			continue
		}

		// Register flag for this field
		if err := l.registerFieldFlag(field, fieldType, prefix); err != nil {
			return err
		}
	}

	return nil
}

// registerFieldFlag registers a flag for a specific field
func (l *Loader) registerFieldFlag(field reflect.Value, fieldType reflect.StructField, prefix string) error {
	flagName := l.getFlagName(fieldType)
	if flagName == "" {
		return nil // No flag for this field
	}

	// Check if flag already registered
	if l.flagSet.Lookup(flagName) != nil {
		return nil // Already registered
	}

	defaultValue := fieldType.Tag.Get("default")
	usage := l.getFlagUsage(fieldType)

	// Register flag based on field type
	switch field.Kind() {
	case reflect.String:
		l.flagSet.String(flagName, defaultValue, usage)
	case reflect.Int:
		defaultInt := 0
		if defaultValue != "" {
			if parsed, err := strconv.Atoi(defaultValue); err == nil {
				defaultInt = parsed
			}
		}
		l.flagSet.Int(flagName, defaultInt, usage)
	case reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			defaultDuration := time.Duration(0)
			if defaultValue != "" {
				if parsed, err := time.ParseDuration(defaultValue); err == nil {
					defaultDuration = parsed
				}
			}
			l.flagSet.Duration(flagName, defaultDuration, usage)
		} else {
			defaultInt64 := int64(0)
			if defaultValue != "" {
				if parsed, err := strconv.ParseInt(defaultValue, 10, 64); err == nil {
					defaultInt64 = parsed
				}
			}
			l.flagSet.Int64(flagName, defaultInt64, usage)
		}
	case reflect.Uint:
		defaultUint := uint(0)
		if defaultValue != "" {
			if parsed, err := strconv.ParseUint(defaultValue, 10, 64); err == nil {
				defaultUint = uint(parsed)
			}
		}
		l.flagSet.Uint(flagName, defaultUint, usage)
	case reflect.Uint64:
		defaultUint64 := uint64(0)
		if defaultValue != "" {
			if parsed, err := strconv.ParseUint(defaultValue, 10, 64); err == nil {
				defaultUint64 = parsed
			}
		}
		l.flagSet.Uint64(flagName, defaultUint64, usage)
	case reflect.Float64:
		defaultFloat := 0.0
		if defaultValue != "" {
			if parsed, err := strconv.ParseFloat(defaultValue, 64); err == nil {
				defaultFloat = parsed
			}
		}
		l.flagSet.Float64(flagName, defaultFloat, usage)
	case reflect.Bool:
		defaultBool := false
		if defaultValue != "" {
			if parsed, err := strconv.ParseBool(defaultValue); err == nil {
				defaultBool = parsed
			}
		}
		l.flagSet.Bool(flagName, defaultBool, usage)
	case reflect.Slice:
		// For slices, use string flag and parse later
		l.flagSet.String(flagName, defaultValue, usage)
	default:
		return fmt.Errorf("unsupported flag type %s for field %s", field.Kind(), fieldType.Name)
	}

	return nil
}

// getFlagUsage generates usage text for a flag
func (l *Loader) getFlagUsage(field reflect.StructField) string {
	if usage := field.Tag.Get("usage"); usage != "" {
		return usage
	}

	if desc := field.Tag.Get("description"); desc != "" {
		return desc
	}

	// Generate default usage
	envKey := field.Tag.Get("env")
	if envKey == "" {
		envKey = toSnakeCase(field.Name)
	}

	usage := fmt.Sprintf("%s (env: %s)", field.Name, strings.ToUpper(envKey))

	if defaultVal := field.Tag.Get("default"); defaultVal != "" {
		usage += fmt.Sprintf(" (default: %s)", defaultVal)
	}

	if field.Tag.Get("required") == "true" {
		usage += " [required]"
	}

	return usage
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

// processField processes a single field
func (l *Loader) processField(field reflect.Value, fieldType reflect.StructField, prefix string) error {
	// Get configuration from struct tags
	envKey := l.getEnvKey(fieldType, prefix)
	flagName := l.getFlagName(fieldType)
	defaultValue := fieldType.Tag.Get("default")
	required := fieldType.Tag.Get("required") == "true"

	// Priority: 1. Flag, 2. Environment, 3. Default
	var value string
	var found bool

	// Check command line flag first
	if flagName != "" {
		if flagValue := l.getFlagValue(flagName); flagValue != "" {
			value = flagValue
			found = true
		}
	}

	// Check environment variable
	if !found && envKey != "" {
		if envValue := os.Getenv(envKey); envValue != "" {
			value = envValue
			found = true
		}
	}

	// Use default value
	if !found && defaultValue != "" {
		value = defaultValue
		found = true
	}

	// Check if required
	if required && !found {
		return fmt.Errorf("required field %s not set", fieldType.Name)
	}

	if found {
		return l.setFieldValue(field, value, fieldType.Name)
	}

	return nil
}

// setFieldValue sets the field value with proper type conversion
func (l *Loader) setFieldValue(field reflect.Value, value, fieldName string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration for %s: %v", fieldName, err)
			}
			field.SetInt(int64(duration))
		} else {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer for %s: %v", fieldName, err)
			}
			field.SetInt(intVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer for %s: %v", fieldName, err)
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float for %s: %v", fieldName, err)
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean for %s: %v", fieldName, err)
		}
		field.SetBool(boolVal)
	case reflect.Slice:
		return l.setSliceValue(field, value, fieldName)
	default:
		return fmt.Errorf("unsupported field type %s for %s", field.Kind(), fieldName)
	}

	return nil
}

// setSliceValue handles slice types
func (l *Loader) setSliceValue(field reflect.Value, value, fieldName string) error {
	if value == "" {
		return nil
	}

	separator := ","
	parts := strings.Split(value, separator)

	slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		elem := slice.Index(i)

		if err := l.setFieldValue(elem, part, fmt.Sprintf("%s[%d]", fieldName, i)); err != nil {
			return err
		}
	}

	field.Set(slice)
	return nil
}

// getEnvKey gets the environment variable key for a field
func (l *Loader) getEnvKey(field reflect.StructField, prefix string) string {
	if envTag := field.Tag.Get("env"); envTag != "" {
		// Support multiple env names: env:"PORT,SERVER_PORT"
		envNames := strings.Split(envTag, ",")
		for _, name := range envNames {
			name = strings.TrimSpace(name)
			if prefix != "" {
				name = prefix + name
			}
			if l.envPrefix != "" {
				name = l.envPrefix + name
			}
			if os.Getenv(name) != "" {
				return name
			}
		}
		// Return first option with prefixes applied
		name := strings.TrimSpace(envNames[0])
		if prefix != "" {
			name = prefix + name
		}
		if l.envPrefix != "" {
			name = l.envPrefix + name
		}
		return name
	}

	// Default: convert field name to UPPER_SNAKE_CASE
	envKey := toSnakeCase(field.Name)
	if prefix != "" {
		envKey = prefix + envKey
	}
	if l.envPrefix != "" {
		envKey = l.envPrefix + envKey
	}
	return strings.ToUpper(envKey)
}

// getFlagName gets the flag name for a field
func (l *Loader) getFlagName(field reflect.StructField) string {
	if flagTag := field.Tag.Get("flag"); flagTag != "" {
		// Support multiple flag names: flag:"port,p"
		flagNames := strings.Split(flagTag, ",")
		return strings.TrimSpace(flagNames[0])
	}

	// Default: convert field name to kebab-case
	return toKebabCase(field.Name)
}

// getFlagValue gets value from command line flags
func (l *Loader) getFlagValue(name string) string {
	if f := l.flagSet.Lookup(name); f != nil {
		return f.Value.String()
	}
	return ""
}

// getNestedPrefix gets the prefix for nested structs
func (l *Loader) getNestedPrefix(field reflect.StructField, currentPrefix string) string {
	if prefixTag := field.Tag.Get("prefix"); prefixTag != "" {
		return currentPrefix + prefixTag
	}
	return currentPrefix + toSnakeCase(field.Name) + "_"
}

// Utility functions
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func toKebabCase(s string) string {
	return strings.ReplaceAll(toSnakeCase(s), "_", "-")
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

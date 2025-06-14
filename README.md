# enfl

A powerful Go library for loading configuration from environment variables, `.env` files, and command-line flags into structs with automatic type conversion and validation.

## Features

- 🔧 **Multiple configuration sources**: `.env` files, environment variables, and command-line flags
- 🎯 **Type-safe**: Supports all Go basic types, slices, and `time.Duration`
- 📋 **Struct tags**: Configure field mapping, defaults, validation, and help text
- 🔄 **Priority system**: Flags override env vars, env vars override `.env` files, `.env` files override defaults
- 🏗️ **Nested structs**: Support for complex configuration structures with prefixes
- ✅ **Validation**: Required field validation and custom error handling
- 🎛️ **Customizable**: Flexible loader options for different use cases

## Installation

```bash
go get github.com/islamghany/enfl
```

## Quick Start

```go
package main

import (
    "fmt"
    "time"
    "github.com/islamghany/enfl"
)

type Config struct {
    Port     int           `env:"PORT" flag:"port" default:"8080" usage:"Server port"`
    Debug    bool          `env:"DEBUG" flag:"debug" default:"false" usage:"Enable debug mode"`
    Timeout  time.Duration `env:"TIMEOUT" flag:"timeout" default:"30s" usage:"Request timeout"`
    Database DatabaseConfig `prefix:"DB_"`
}

type DatabaseConfig struct {
    Host     string `env:"HOST" default:"localhost" usage:"Database host"`
    Port     int    `env:"PORT" default:"5432" usage:"Database port"`
    Name     string `env:"NAME" required:"true" usage:"Database name"`
    Username string `env:"USERNAME" required:"true" usage:"Database username"`
    Password string `env:"PASSWORD" required:"true" usage:"Database password"`
}

func main() {
    var cfg Config
    if err := enfl.Load(&cfg); err != nil {
        panic(err)
    }
    fmt.Printf("Config: %+v\n", cfg)
}
```

## Struct Tag Reference

| Tag        | Description                     | Example                                  |
| ---------- | ------------------------------- | ---------------------------------------- |
| `env`      | Environment variable name(s)    | `env:"PORT"` or `env:"PORT,SERVER_PORT"` |
| `flag`     | Command-line flag name(s)       | `flag:"port"` or `flag:"port,p"`         |
| `default`  | Default value if not set        | `default:"8080"`                         |
| `usage`    | Help text for flags             | `usage:"Port to listen on"`              |
| `required` | Field must be provided          | `required:"true"`                        |
| `prefix`   | Prefix for nested struct fields | `prefix:"DB_"`                           |

## Complete Feature Examples

### 1. Basic Configuration Loading

```go
type Config struct {
    AppName string `env:"APP_NAME" flag:"name" default:"myapp"`
    Port    int    `env:"PORT" flag:"port" default:"8080"`
    Debug   bool   `env:"DEBUG" flag:"debug" default:"false"`
}

var cfg Config
err := enfl.Load(&cfg)
```

**Usage:**

```bash
# Via environment variables
export APP_NAME=webapp PORT=9000 DEBUG=true
go run main.go

# Via command-line flags
go run main.go -name=webapp -port=9000 -debug=true

# Via .env file
echo "APP_NAME=webapp" > .env
echo "PORT=9000" >> .env
echo "DEBUG=true" >> .env
go run main.go
```

### 2. All Supported Types

```go
type AllTypesConfig struct {
    // Basic types
    StringVal  string  `env:"STRING_VAL" default:"hello"`
    IntVal     int     `env:"INT_VAL" default:"42"`
    Int8Val    int8    `env:"INT8_VAL" default:"8"`
    Int16Val   int16   `env:"INT16_VAL" default:"16"`
    Int32Val   int32   `env:"INT32_VAL" default:"32"`
    Int64Val   int64   `env:"INT64_VAL" default:"64"`
    UintVal    uint    `env:"UINT_VAL" default:"42"`
    Uint8Val   uint8   `env:"UINT8_VAL" default:"8"`
    Uint16Val  uint16  `env:"UINT16_VAL" default:"16"`
    Uint32Val  uint32  `env:"UINT32_VAL" default:"32"`
    Uint64Val  uint64  `env:"UINT64_VAL" default:"64"`
    Float32Val float32 `env:"FLOAT32_VAL" default:"3.14"`
    Float64Val float64 `env:"FLOAT64_VAL" default:"2.718"`
    BoolVal    bool    `env:"BOOL_VAL" default:"true"`

    // Duration
    Timeout time.Duration `env:"TIMEOUT" default:"5m30s"`

    // Slices (comma-separated)
    StringSlice []string  `env:"STRING_SLICE" default:"a,b,c"`
    IntSlice    []int     `env:"INT_SLICE" default:"1,2,3"`
    FloatSlice  []float64 `env:"FLOAT_SLICE" default:"1.1,2.2,3.3"`
}
```

### 3. Nested Configuration with Prefixes

```go
type ServerConfig struct {
    Host string `env:"HOST" default:"localhost"`
    Port int    `env:"PORT" default:"8080"`
    TLS  bool   `env:"TLS" default:"false"`
}

type DatabaseConfig struct {
    Driver   string `env:"DRIVER" default:"postgres"`
    Host     string `env:"HOST" default:"localhost"`
    Port     int    `env:"PORT" default:"5432"`
    Database string `env:"DATABASE" required:"true"`
    Username string `env:"USERNAME" required:"true"`
    Password string `env:"PASSWORD" required:"true"`
}

type Config struct {
    AppName  string         `env:"APP_NAME" default:"myapp"`
    Server   ServerConfig   `prefix:"SERVER_"`
    Database DatabaseConfig `prefix:"DB_"`
}
```

**Environment variables:**

```bash
APP_NAME=webapp
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_TLS=true
DB_DRIVER=postgres
DB_HOST=db.example.com
DB_PORT=5432
DB_DATABASE=myapp
DB_USERNAME=user
DB_PASSWORD=secret
```

### 4. Required Fields and Validation

```go
type Config struct {
    APIKey    string `env:"API_KEY" required:"true" usage:"API key for external service"`
    SecretKey string `env:"SECRET_KEY" required:"true" usage:"Secret key for encryption"`
    Optional  string `env:"OPTIONAL" default:"default_value"`
}

var cfg Config
if err := enfl.Load(&cfg); err != nil {
    fmt.Printf("Configuration error: %v\n", err)
    // Will error if API_KEY or SECRET_KEY are not provided
}
```

### 5. Multiple Names for Same Field

```go
type Config struct {
    Port int `env:"PORT,SERVER_PORT" flag:"port,p" default:"8080"`
}
```

**Usage:**

```bash
# Any of these work:
export PORT=9000
export SERVER_PORT=9000
go run main.go -port=9000
go run main.go -p=9000
```

### 6. Custom Loader Configuration

```go
// Custom loader with specific options
loader := enfl.NewLoader(
    enfl.WithEnvPrefix("MYAPP_"),           // Add prefix to all env vars
    enfl.WithEnvFiles(".env", ".env.local"), // Load specific .env files
    enfl.WithFailOnError(false),            // Continue on errors
)

var cfg Config
err := loader.Load(&cfg)
```

### 7. Working with .env Files

Create `.env` file:

```env
# Application settings
APP_NAME=mywebapp
PORT=8080
DEBUG=true

# Database settings
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp
DB_USER=admin
DB_PASSWORD=secret123

# Array values (comma-separated)
ALLOWED_HOSTS=localhost,127.0.0.1,example.com
FEATURE_FLAGS=auth,logging,metrics
```

### 8. Command-line Flag Integration

```go
type Config struct {
    Verbose bool   `flag:"verbose,v" usage:"Enable verbose logging"`
    Config  string `flag:"config,c" usage:"Path to config file"`
    Port    int    `flag:"port,p" default:"8080" usage:"Port to listen on"`
}

var cfg Config
enfl.Load(&cfg)

// Generates help automatically:
// go run main.go -help
```

### 9. Priority System Example

```go
type Config struct {
    Port int `env:"PORT" flag:"port" default:"8080"`
}
```

**Priority order (highest to lowest):**

1. Command-line flags: `go run main.go -port=9000`
2. Environment variables: `export PORT=8090`
3. `.env` file: `PORT=8085`
4. Default value: `8080`

### 10. Complex Real-world Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/islamghany/enfl"
)

type LogConfig struct {
    Level  string `env:"LEVEL" default:"info" usage:"Log level (debug, info, warn, error)"`
    Format string `env:"FORMAT" default:"json" usage:"Log format (json, text)"`
    File   string `env:"FILE" usage:"Log file path (empty for stdout)"`
}

type DatabaseConfig struct {
    Driver          string        `env:"DRIVER" default:"postgres"`
    Host            string        `env:"HOST" default:"localhost"`
    Port            int           `env:"PORT" default:"5432"`
    Name            string        `env:"NAME" required:"true"`
    Username        string        `env:"USERNAME" required:"true"`
    Password        string        `env:"PASSWORD" required:"true"`
    MaxConnections  int           `env:"MAX_CONNECTIONS" default:"10"`
    ConnectTimeout  time.Duration `env:"CONNECT_TIMEOUT" default:"10s"`
    SSLMode         string        `env:"SSL_MODE" default:"prefer"`
}

type RedisConfig struct {
    Host     string        `env:"HOST" default:"localhost"`
    Port     int           `env:"PORT" default:"6379"`
    Password string        `env:"PASSWORD"`
    DB       int           `env:"DB" default:"0"`
    Timeout  time.Duration `env:"TIMEOUT" default:"5s"`
}

type ServerConfig struct {
    Host           string        `env:"HOST" default:"0.0.0.0"`
    Port           int           `env:"PORT" default:"8080"`
    ReadTimeout    time.Duration `env:"READ_TIMEOUT" default:"30s"`
    WriteTimeout   time.Duration `env:"WRITE_TIMEOUT" default:"30s"`
    MaxHeaderBytes int           `env:"MAX_HEADER_BYTES" default:"1048576"`
    TLS            TLSConfig     `prefix:"TLS_"`
}

type TLSConfig struct {
    Enabled  bool   `env:"ENABLED" default:"false"`
    CertFile string `env:"CERT_FILE"`
    KeyFile  string `env:"KEY_FILE"`
}

type Config struct {
    AppName     string         `env:"APP_NAME" flag:"name" default:"webapp" usage:"Application name"`
    Environment string         `env:"ENVIRONMENT" flag:"env" default:"development" usage:"Environment (development, staging, production)"`
    Debug       bool           `env:"DEBUG" flag:"debug" default:"false" usage:"Enable debug mode"`

    Server   ServerConfig   `prefix:"SERVER_"`
    Database DatabaseConfig `prefix:"DB_"`
    Redis    RedisConfig    `prefix:"REDIS_"`
    Log      LogConfig      `prefix:"LOG_"`

    // Feature flags
    Features []string `env:"FEATURES" flag:"features" usage:"Comma-separated list of enabled features"`

    // External APIs
    APIKeys map[string]string `env:"API_KEYS"` // Note: maps need custom handling
}

func main() {
    var cfg Config

    // Load with custom options
    loader := enfl.NewLoader(
        enfl.WithEnvFiles(".env", ".env.local"),
        enfl.WithFailOnError(true),
    )

    if err := loader.Load(&cfg); err != nil {
        fmt.Printf("Failed to load configuration: %v\n", err)
        return
    }

    fmt.Printf("Configuration loaded successfully:\n")
    fmt.Printf("App: %s (%s)\n", cfg.AppName, cfg.Environment)
    fmt.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
    fmt.Printf("Database: %s@%s:%d/%s\n", cfg.Database.Username, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
    fmt.Printf("Redis: %s:%d\n", cfg.Redis.Host, cfg.Redis.Port)
    fmt.Printf("Features: %v\n", cfg.Features)
}
```

**Example `.env` file for the above:**

```env
APP_NAME=MyWebApp
ENVIRONMENT=production
DEBUG=false

SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_TLS_ENABLED=true
SERVER_TLS_CERT_FILE=/etc/ssl/certs/app.crt
SERVER_TLS_KEY_FILE=/etc/ssl/private/app.key

DB_HOST=db.example.com
DB_PORT=5432
DB_NAME=myapp_prod
DB_USERNAME=appuser
DB_PASSWORD=securepassword123
DB_MAX_CONNECTIONS=20

REDIS_HOST=redis.example.com
REDIS_PORT=6379
REDIS_PASSWORD=redispass

LOG_LEVEL=info
LOG_FORMAT=json
LOG_FILE=/var/log/myapp.log

FEATURES=auth,billing,analytics,notifications
```

## API Reference

### Functions

- `Load(ptr interface{}) error` - Load configuration using default loader
- `NewLoader(options ...Option) *Loader` - Create custom loader with options

### Loader Options

- `WithEnvPrefix(prefix string)` - Add prefix to all environment variables
- `WithEnvFiles(files ...string)` - Specify .env files to load
- `WithFailOnError(fail bool)` - Control error handling behavior

## Error Handling

The library provides detailed error messages for common issues:

- Missing required fields
- Type conversion errors
- Invalid default values
- File loading errors

```go
if err := enfl.Load(&cfg); err != nil {
    fmt.Printf("Configuration error: %v\n", err)
    // Handle specific error types if needed
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License

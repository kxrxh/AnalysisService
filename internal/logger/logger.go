package logger

import (
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	Logger zerolog.Logger
)

// LogLevel represents available log levels
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Config holds logger configuration
type Config struct {
	Level      LogLevel
	UseConsole bool
	Output     io.Writer
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:      LevelInfo,
		UseConsole: true,
		Output:     os.Stdout,
	}
}

func init() {
	// Initialize with default configuration
	Initialize(DefaultConfig())
}

// Initialize sets up the logger with the provided configuration
func Initialize(config Config) {
	zerolog.TimeFieldFormat = time.RFC3339

	// Set global log level
	setGlobalLevel(config.Level)

	// Configure output
	var output io.Writer = config.Output
	if config.UseConsole {
		output = zerolog.ConsoleWriter{
			Out:        config.Output,
			TimeFormat: time.RFC3339,
		}
	}

	Logger = log.Output(output)
}

// SetLevel changes the logging level dynamically
func SetLevel(level LogLevel) {
	setGlobalLevel(level)
	Logger.Info().Str("level", string(level)).Msg("Log level changed")
}

// SetLevelFromString sets the log level from a string value
func SetLevelFromString(levelStr string) error {
	level, err := parseLogLevel(levelStr)
	if err != nil {
		return err
	}
	SetLevel(level)
	return nil
}

// GetLevel returns the current log level
func GetLevel() LogLevel {
	switch zerolog.GlobalLevel() {
	case zerolog.DebugLevel:
		return LevelDebug
	case zerolog.InfoLevel:
		return LevelInfo
	case zerolog.WarnLevel:
		return LevelWarn
	case zerolog.ErrorLevel:
		return LevelError
	default:
		return LevelInfo
	}
}

// setGlobalLevel converts LogLevel to zerolog level and sets it globally
func setGlobalLevel(level LogLevel) {
	var zeroLevel zerolog.Level

	switch level {
	case LevelDebug:
		zeroLevel = zerolog.DebugLevel
	case LevelInfo:
		zeroLevel = zerolog.InfoLevel
	case LevelWarn:
		zeroLevel = zerolog.WarnLevel
	case LevelError:
		zeroLevel = zerolog.ErrorLevel
	default:
		zeroLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(zeroLevel)
}

// parseLogLevel converts string to LogLevel
func parseLogLevel(levelStr string) (LogLevel, error) {
	switch strings.ToLower(levelStr) {
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	default:
		return LevelInfo, errors.New("invalid log level")
	}
}

// EnableConsoleOutput switches to console-friendly output
func EnableConsoleOutput() {
	config := Config{
		Level:      GetLevel(),
		UseConsole: true,
		Output:     os.Stdout,
	}
	Initialize(config)
}

// EnableJSONOutput switches to JSON output (for production)
func EnableJSONOutput() {
	config := Config{
		Level:      GetLevel(),
		UseConsole: false,
		Output:     os.Stdout,
	}
	Initialize(config)
}

// InitLogger initializes the global logger with pretty printing for development
func InitLogger() {
	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339

	// Pretty print (human-readable) logs by default. This follows the user preference to always have
	// readable logs even in production. You can disable pretty output by setting
	// LOG_CONSOLE_OUTPUT=false|0|json to force JSON output.
	useConsoleOutput := true
	if v := strings.ToLower(os.Getenv("LOG_CONSOLE_OUTPUT")); v == "false" || v == "0" || v == "json" {
		useConsoleOutput = false
	}

	if useConsoleOutput {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		log.Logger = log.Output(consoleWriter)
		Logger = log.Output(consoleWriter)
	} else {
		Logger = log.Output(os.Stdout)
	}

	// Set log level
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.DebugLevel) // Changed default to DebugLevel
	}
}

// GetLogger returns a logger instance with the specified component name
func GetLogger(component string) zerolog.Logger {
	return Logger.With().Str("component", component).Logger()
}

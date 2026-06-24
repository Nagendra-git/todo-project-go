// Package config loads application settings from a Java-style
// .properties file, with environment variables taking precedence
// over file values, so deployments can override settings without
// editing the file.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application settings in a typed, validated form.
type Config struct {
	Mongo  MongoConfig
	Server ServerConfig
	CORS   CORSConfig
}

type MongoConfig struct {
	URI        string
	Database   string
	Collection string
	Timeout    time.Duration
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type CORSConfig struct {
	AllowedOrigin string
}

// Load reads the properties file at path, applies environment
// variable overrides, and returns a validated Config.
func Load(path string) (*Config, error) {
	raw, err := readProperties(path)
	if err != nil {
		// Missing file isn't fatal — fall back to defaults/env vars.
		raw = map[string]string{}
	}

	get := func(key, def string) string {
		envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		if v := os.Getenv(envKey); v != "" {
			return v
		}
		if v, ok := raw[key]; ok && v != "" {
			return v
		}
		return def
	}

	getInt := func(key string, def int) int {
		v := get(key, "")
		if v == "" {
			return def
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			return def
		}
		return n
	}

	cfg := &Config{
		Mongo: MongoConfig{
			URI:        get("mongo.uri", "mongodb://localhost:27017"),
			Database:   get("mongo.database", "tododb"),
			Collection: get("mongo.collection", "todos"),
			Timeout:    time.Duration(getInt("mongo.timeout.seconds", 10)) * time.Second,
		},
		Server: ServerConfig{
			Port:         get("server.port", "8080"),
			ReadTimeout:  time.Duration(getInt("server.read_timeout.seconds", 5)) * time.Second,
			WriteTimeout: time.Duration(getInt("server.write_timeout.seconds", 10)) * time.Second,
		},
		CORS: CORSConfig{
			AllowedOrigin: get("cors.allowed_origin", "*"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Mongo.URI == "" {
		return fmt.Errorf("mongo.uri must not be empty")
	}
	if c.Mongo.Database == "" {
		return fmt.Errorf("mongo.database must not be empty")
	}
	if c.Server.Port == "" {
		return fmt.Errorf("server.port must not be empty")
	}
	return nil
}

// readProperties parses a simple key=value file, skipping blank
// lines and comments (# or !), as used by Java's .properties format.
func readProperties(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		values[key] = val
	}
	return values, scanner.Err()
}

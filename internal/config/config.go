package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

type DBConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

type JWTConfig struct {
	SecretKey  string `json:"secretKey"`
	Issuer     string `json:"issuer"`
	Audience   string `json:"audience"`
	CookieName string `json:"cookieName"`
	ExpireDays int    `json:"expireDays"`
}

type Config struct {
	Env  string    `json:"env"`
	Port string    `json:"port"`
	JWT  JWTConfig `json:"jwt"`
	DB   DBConfig  `json:"db"`
}

func Load() (*Config, error) {
	var cfg *Config
	var err error
	switch runtime.GOOS {
	case "windows":
		cfg, err = loadConfigJSON()
	case "linux":
		cfg, err = loadConfigEnv()
	default:
		return nil, fmt.Errorf("config loading error")
	}

	return cfg, err
}

func loadConfigEnv() (*Config, error) {
	var err error
	getEnv := func(key string) string {
		if err != nil {
			return ""
		}
		val, ok := os.LookupEnv(key)
		if !ok {
			err = fmt.Errorf("missing environment variable: %v", key)
			return ""
		}
		return val
	}

	cfg := &Config{
		Env:  getEnv("ENV"),
		Port: getEnv("PORT"),
		JWT: JWTConfig{
			SecretKey:  getEnv("JWT_SECRET"),
			Issuer:     getEnv("JWT_ISSUER"),
			Audience:   getEnv("JWT_AUDIENCE"),
			CookieName: getEnv("JWT_COOKIE"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST"),
			User:     getEnv("DB_USER"),
			Password: getEnv("DB_PASSWORD"),
			DBName:   getEnv("DB_NAME"),
		},
	}

	expireDays, err := strconv.Atoi(getEnv("JWT_EXPIREDAYS"))

	if err != nil {
		return nil, err
	}

	cfg.JWT.ExpireDays = expireDays

	return cfg, nil
}

func loadConfigJSON() (*Config, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	currDir := filepath.Dir(exePath)
	file, err := os.Open(filepath.Join(currDir, "config.json"))
	if err != nil {
		return nil, fmt.Errorf("json config file oppening error: %v", err)
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("json config file reading error: %v", err)
	}

	return &cfg, nil
}

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

type CredConfig struct {
	LoginChars     string `json:"loginChars"`
	MaxLoginLength int    `json:"maxLoginLength"`
	MinLoginLength int    `json:"minLoginLength"`
	PwdChars       string `json:"passwordChars"`
	MaxPwdLength   int    `json:"maxPasswordLength"`
	MinPwdLength   int    `json:"minPasswordLength"`
}

type Config struct {
	Env  string     `json:"env"`
	Port string     `json:"port"`
	JWT  JWTConfig  `json:"jwt"`
	DB   DBConfig   `json:"db"`
	Cred CredConfig `json:"cred"`
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
	getEnvString := func(key string) string {
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

	getEnvInt := func(key string) int {
		if err != nil {
			return 0
		}
		val, ok := os.LookupEnv(key)
		if !ok {
			err = fmt.Errorf("missing environment variable: %v", key)
			return 0
		}

		valInt, err := strconv.Atoi(val)
		if err != nil {
			err = fmt.Errorf("environment variable is not integer: %v", val)
			return 0
		}

		return valInt
	}

	cfg := &Config{
		Env:  getEnvString("ENV"),
		Port: getEnvString("PORT"),
		JWT: JWTConfig{
			SecretKey:  getEnvString("JWT_SECRET"),
			Issuer:     getEnvString("JWT_ISSUER"),
			Audience:   getEnvString("JWT_AUDIENCE"),
			CookieName: getEnvString("JWT_COOKIE"),
			ExpireDays: getEnvInt("JWT_EXPIREDAYS"),
		},
		DB: DBConfig{
			Host:     getEnvString("DB_HOST"),
			User:     getEnvString("DB_USER"),
			Password: getEnvString("DB_PASSWORD"),
			DBName:   getEnvString("DB_NAME"),
		},
		Cred: CredConfig{
			LoginChars:     getEnvString("CRED_LOGIN_CHARS"),
			MaxLoginLength: getEnvInt("CRED_MAX_LOGIN_LENGTH"),
			MinLoginLength: getEnvInt("CRED_MIN_LOGIN_LENGTH"),
			PwdChars:       getEnvString("CRED_PASSWORD_CHARS"),
			MaxPwdLength:   getEnvInt("CRED_MAX_PASSWORD_LENGTH"),
			MinPwdLength:   getEnvInt("CRED_MIN_PASSWORD_LENGTH"),
		},
	}

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

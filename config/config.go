package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	ServerPort string `validate:"required,len=4"`

	AmadeusAPIKey    string `validate:"required,min=25"`
	AmadeusAPISecret string `validate:"required,min=10"`
	AmadeusBaseURL   string `validate:"required,min=15"`

	SkyScannerRapidAPIKey string `validate:"required,min=25"`
	SkyScannerBaseURL     string `validate:"required,min=15"`

	AppBaseURL string `validate:"required,url"`
	AppEnv     string `validate:"required,min=5"`

	ClientTimeout time.Duration `validate:"required,len=3"`
}

func Load() (*Config, error) {
	env := getEnvOrSecret("APP_ENV", "")
	log.Println("the current environment is: ", env)

	clientTimeout, _ := time.ParseDuration(getEnvOrSecret("CLIENT_TIMEOUT", "10s"))

	c := Config{
		ServerPort:            getEnvOrSecret("SERVER_PORT", ""),
		AmadeusAPIKey:         getEnvOrSecret("AMADEUS_API_KEY", "SECRET"),
		AmadeusAPISecret:      getEnvOrSecret("AMADEUS_API_SECRET", "SECRET"),
		AmadeusBaseURL:        getEnvOrSecret("AMADEUS_BASE_URL", ""),
		SkyScannerRapidAPIKey: getEnvOrSecret("SKYSCANNER_API_KEY", "SECRET"),
		SkyScannerBaseURL:     getEnvOrSecret("SKYSCANNER_BASE_URL", ""),
		AppBaseURL:            getEnvOrSecret("APP_BASE_URL", ""),
		AppEnv:                getEnvOrSecret("APP_ENV", ""),
		ClientTimeout:         clientTimeout,
	}
	if err := validate(c); err != nil {
		return nil, err
	}
	return &c, nil
}

func getEnvOrSecret(key, defaultValue string) string {
	// if it's not a secret info the variable should be in the environment
	if defaultValue != "SECRET" {
		if value, exists := os.LookupEnv(key); exists {
			return value
		}
	}

	// intent to read docker secret
	if secret, err := readDockerSecret(key); err == nil && secret != "" {
		return secret
	}

	return defaultValue
}

func readDockerSecret(secretName string) (string, error) {
	secretPath := fmt.Sprintf("/run/secrets/%s", secretName)

	data, err := os.ReadFile(secretPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // does not exist then not return error, the specific error check occurs in validate
		}
		return "", fmt.Errorf("error reading secret %s: %v", secretName, err)
	}

	return strings.TrimSpace(string(data)), nil
}

func validate(config Config) error {
	validate := validator.New()

	if err := validate.Struct(config); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return fmt.Errorf("The variable %s is not vaild: %s\n", err.Tag(), err.Field())
		}
	}
	return nil
}

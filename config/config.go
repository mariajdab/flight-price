package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string `validate:"required,len=4"`

	AmadeusAPIKey    string `validate:"required,min=25"`
	AmadeusAPISecret string `validate:"required,min=10"`
	AmadeusBaseURL   string `validate:"required,min=15"`

	SkyScannerRapidAPIKey string `validate:"required,min=25"`
	SkyScannerBaseURL     string `validate:"required,min=15"`

	ClientTimeout time.Duration `validate:"required,len=3"`
}

func Load() (*Config, error) {
	// First intent to charge the variable from the .env (this is only for testing, the code use docker secret for production)
	_ = godotenv.Load(".env") // ignore the error

	env := getEnvOrSecret("APP_ENV", "development")
	log.Println("the current environment is: ", env)

	clientTimeout, _ := time.ParseDuration(getEnvOrSecret("CLIENT_TIMEOUT", "10s"))

	c := Config{
		ServerPort:            getEnvOrSecret("SERVER_PORT", "8080"),
		AmadeusAPIKey:         getEnvOrSecret("AMADEUS_API_KEY", ""),
		AmadeusAPISecret:      getEnvOrSecret("AMADEUS_API_SECRET", ""),
		AmadeusBaseURL:        getEnvOrSecret("AMADEUS_BASE_URL", ""),
		SkyScannerRapidAPIKey: getEnvOrSecret("SKYSCANNER_API_KEY", ""),
		SkyScannerBaseURL:     getEnvOrSecret("SKYSCANNER_BASE_URL", ""),
		ClientTimeout:         clientTimeout,
	}
	err := validate(c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func getEnvOrSecret(key, defaultValue string) string {
	// Op 1. The code intent to read docker secret (production)
	if secret, err := readDockerSecret(key); err == nil && secret != "" {
		return secret
	}

	// Op 2. Search the variable from env (useful for testing propose)
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	// Op 3. Use default value
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

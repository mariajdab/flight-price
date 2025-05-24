package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
)

const dockerSecretPathPrefix = "/run/secrets"

type Config struct {
	ServerPort string `validate:"required,len=4"`

	AmadeusAPIKey    string `validate:"required,min=25"`
	AmadeusAPISecret string `validate:"required,min=10"`
	AmadeusBaseURL   string `validate:"required,min=15"`

	SkyRapidAPIKey  string `validate:"required,min=25"`
	SkyRapidBaseURL string `validate:"required,min=15"`

	GoogleFlightRapidAPIKey  string `validate:"required,min=25"`
	GoogleFlightRapidBaseURL string `validate:"required,min=15"`

	AppBaseURL string `validate:"required,url"`
	AppEnv     string `validate:"required,min=5"`

	ClientTimeout time.Duration `validate:"required"`
}

func Load() (*Config, error) {
	env := getEnvOrFail("APP_ENV")
	log.Println("the current environment is: ", env)

	clientTimeout, err := time.ParseDuration(getEnvOrFail("CLIENT_TIMEOUT"))
	if err != nil {
		return nil, err
	}

	amadeusAPIKey, err := os.ReadFile(filepath.Join(
		dockerSecretPathPrefix,
		getEnvOrFail("AMADEUS_API_KEY"),
	))
	if err != nil {
		return nil, err
	}
	amadeusAPISecret, err := os.ReadFile(filepath.Join(
		dockerSecretPathPrefix,
		getEnvOrFail("AMADEUS_API_SECRET"),
	))
	if err != nil {
		return nil, err
	}
	skyRapidAPIKey, err := os.ReadFile(filepath.Join(
		dockerSecretPathPrefix,
		getEnvOrFail("SKY_RAPID_API_KEY"),
	))
	if err != nil {
		return nil, err
	}
	googleFlightAPIKey, err := os.ReadFile(filepath.Join(
		dockerSecretPathPrefix,
		getEnvOrFail("GOOGLE_FLIGHT_RAPID_API_KEY"),
	))
	if err != nil {
		return nil, err
	}

	c := Config{
		AppEnv:                   getEnvOrFail("APP_ENV"),
		ServerPort:               getEnvOrFail("SERVER_PORT"),
		AppBaseURL:               getEnvOrFail("APP_BASE_URL"),
		AmadeusBaseURL:           getEnvOrFail("AMADEUS_BASE_URL"),
		SkyRapidBaseURL:          getEnvOrFail("SKY_RAPID_BASE_URL"),
		GoogleFlightRapidBaseURL: getEnvOrFail("GOOGLE_FLIGHT_RAPID_BASE_URL"),
		AmadeusAPIKey:            string(amadeusAPIKey),
		AmadeusAPISecret:         string(amadeusAPISecret),
		SkyRapidAPIKey:           string(skyRapidAPIKey),
		GoogleFlightRapidAPIKey:  string(googleFlightAPIKey),
		ClientTimeout:            clientTimeout,
	}
	if err := validate(c); err != nil {
		return nil, err
	}
	return &c, nil
}

func getEnvOrFail(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	} else {
		panic(fmt.Errorf("could not find env: %s", key))
	}
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

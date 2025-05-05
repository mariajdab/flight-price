package config

import (
	"fmt"
	"log"
	"os"
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

	ClientTimeout time.Duration `validate:"required"`
}

func Load() (*Config, error) {
	env := getEnvOrFail("APP_ENV")
	log.Println("the current environment is: ", env)

	clientTimeout, err := time.ParseDuration(getEnvOrFail("CLIENT_TIMEOUT"))
	if err != nil {
		return nil, err
	}

	c := Config{
		ServerPort:            getEnvOrFail("SERVER_PORT"),
		AmadeusAPIKey:         getEnvOrFail("AMADEUS_API_KEY"),
		AmadeusAPISecret:      getEnvOrFail("AMADEUS_API_SECRET"),
		AmadeusBaseURL:        getEnvOrFail("AMADEUS_BASE_URL"),
		SkyScannerRapidAPIKey: getEnvOrFail("SKYSCANNER_API_KEY"),
		SkyScannerBaseURL:     getEnvOrFail("SKYSCANNER_BASE_URL"),
		AppBaseURL:            getEnvOrFail("APP_BASE_URL"),
		AppEnv:                getEnvOrFail("APP_ENV"),
		ClientTimeout:         clientTimeout,
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

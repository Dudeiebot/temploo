package config

import (
	"errors"
	"os"
)

var AppConfig App

type App struct {
	AppName         string
	AppKey          string
	AsynqmonService string
	AppHost         string
	ApiHost         string
}

func GetApiHost() string {
	if AppConfig.ApiHost == "" {
		return "*"
	}
	return AppConfig.ApiHost
}

func loadAppEnv() error {
	appKey, exists := os.LookupEnv("APP_KEY")
	if !exists {
		return errors.New("APP_KEY is not set")
	}

	appName, exists := os.LookupEnv("APP_NAME")
	if !exists {
		return errors.New("APP_NAME is not set")
	}

	asynqmonService, exists := os.LookupEnv("ASYNQMON_SERVICE")
	if !exists {
		return errors.New("ASYNQMON_SERVICE is not set")
	}

	appHost, exists := os.LookupEnv("APP_HOST")
	if !exists {
		return errors.New("APP_HOST is not set in .env")
	}

	apiHost, exists := os.LookupEnv("API_HOST")
	if !exists {
		return errors.New("API_HOST is not set in .env")
	}

	AppConfig = App{
		AppName:         appName,
		AppKey:          appKey,
		AsynqmonService: asynqmonService,
		AppHost:         appHost,
		ApiHost:         apiHost,
	}

	return nil
}

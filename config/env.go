package config

import "github.com/joho/godotenv"

func LoadEnvironmentVariable() error {
	_ = godotenv.Load()

	err := loadAppEnv()
	if err != nil {
		return err
	}

	err = loadDbEnv()
	if err != nil {
		return err
	}

	err = loadMailEnv()
	if err != nil {
		return err
	}

	return nil
}

package config

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DbConfig DB
	PostDb   *gorm.DB
	Redis    *redis.Client
)

type DB struct {
	DBName      string
	DBPassword  string
	DBUsername  string
	DBPort      string
	DBHost      string
	RedisHost   string
	RedisPass   string
	RedisPort   string
	RedisUser   string
	RedisScheme string
	RedisAddr   string
}

func loadDbEnv() error {
	dBHost, exists := os.LookupEnv("DB_HOST")
	if !exists {
		return errors.New("DB_HOST not in .env")
	}

	dbName, exists := os.LookupEnv("DB_NAME")
	if !exists {
		return errors.New("DB_NAME not in .env")
	}

	dBPassword, exists := os.LookupEnv("DB_PASSWORD")
	if !exists {
		return errors.New("DB_PASSWORD not in .env")
	}

	dBUsername, exists := os.LookupEnv("DB_USERNAME")
	if !exists {
		return errors.New("DB_USERNAME not in .env")
	}

	dBPort, exists := os.LookupEnv("DB_PORT")
	if !exists {
		return errors.New("DB_PORT not in .env")
	}

	redisHost, exists := os.LookupEnv("REDIS_HOST")
	if !exists {
		return errors.New("REDIS_HOST not in .env")
	}

	redisPass, exists := os.LookupEnv("REDIS_PASS")
	if !exists {
		return errors.New("REDIS_PASS not in .env")
	}

	redisPort, exists := os.LookupEnv("REDIS_PORT")
	if !exists {
		return errors.New("REDIS_PORT not in .env")
	}

	redisUser, exists := os.LookupEnv("REDIS_USER")
	if !exists {
		return errors.New("REDIS_USER not in .env")
	}

	redisScheme, exists := os.LookupEnv("REDIS_SCHEME")
	if !exists {
		return errors.New("REDIS_SCHEME not in .env")
	}

	DbConfig = DB{
		DBName:      dbName,
		DBPassword:  dBPassword,
		DBUsername:  dBUsername,
		DBPort:      dBPort,
		DBHost:      dBHost,
		RedisHost:   redisHost,
		RedisPass:   redisPass,
		RedisPort:   redisPort,
		RedisUser:   redisUser,
		RedisScheme: redisScheme,
		RedisAddr:   fmt.Sprintf("%s:%s", redisHost, redisPort),
	}

	return nil
}

func ConnectPostGres(cfg *DB) error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUsername, cfg.DBPassword, cfg.DBName,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	PostDb = db

	return nil
}

func ConnectRedis(cfg *DB) error {
	var client *redis.Client

	if cfg.RedisScheme == "tls" {
		client = redis.NewClient(&redis.Options{
			Addr:       fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
			Password:   cfg.RedisPass,
			DB:         0,
			TLSConfig:  &tls.Config{},
			MaxRetries: 3,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:       fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
			Password:   cfg.RedisPass,
			DB:         0,
			MaxRetries: 3,
		})
	}

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return err
	}

	Redis = client

	connectionInfo := fmt.Sprintf(
		"%s:%s, %s, %s",
		DbConfig.RedisHost,
		DbConfig.RedisPort,
		DbConfig.RedisUser,
		DbConfig.RedisPass,
	)
	fmt.Println("connected to redis", connectionInfo)
	return nil
}

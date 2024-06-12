package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Configuration struct {
	ServerAddr         string
	MongoConn          string `mapstructure:"MONGO_CONN"`
	MongoDbName        string `mapstructure:"MONGO_DB_NAME"`
	AwsS3Bucket        string `mapstructure:"AMAZON_S3BUCKET"`
	AwsRegion          string `mapstructure:"AMAZON_REGION"`
	AwsAccessKeyID     string `mapstructure:"AMAZON_ACCESS_KEY_ID"`
	AwsSecretAccessKey string `mapstructure:"AMAZON_SECRET_ACCESS_KEY"`
}

var Env string
var Config Configuration

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		if value == "" {
			return fallback
		}
		return value
	}
	return fallback
}

func LoadEnv() (config Configuration, err error) {
	env := GetEnv("ENV", "local")
	godotenv.Load("../.env")

	config = Configuration{
		ServerAddr:         os.Getenv("SERVER_ADDR"),
		MongoConn:          os.Getenv("MONGO_CONN"),
		MongoDbName:        os.Getenv("MONGO_DB_NAME"),
		AwsS3Bucket:        os.Getenv("AMAZON_S3BUCKET"),
		AwsRegion:          os.Getenv("AMAZON_REGION"),
		AwsAccessKeyID:     os.Getenv("AMAZON_ACCESS_KEY_ID"),
		AwsSecretAccessKey: os.Getenv("AMAZON_SECRET_ACCESS_KEY"),
	}

	Env = env
	Config = config

	return config, nil
}

func GetEnvFromEnv(env string) (envValue string) {
	envValue, found := os.LookupEnv(env)

	if !found {
		err := godotenv.Load("../.env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}
		envValue = os.Getenv(env)
		return envValue
	} else {
		return envValue
	}
}

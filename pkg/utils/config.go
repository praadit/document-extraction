package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Configuration struct {
	ServerAddr         string `mapstructure:"SERVER_ADDR"`
	MongoConn          string `mapstructure:"MONGO_CONN"`
	MongoDbName        string `mapstructure:"MONGO_DB_NAME"`
	AwsS3Bucket        string `mapstructure:"AWS_S3BUCKET"`
	AwsBucketRegion    string `mapstructure:"AWS_BUCKET_REGION"`
	AwsBedrockRegion   string `mapstructure:"AWS_BEDROCK_REGION"`
	AwsAccessKeyID     string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AwsSecretAccessKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	ChromaAddress      string `mapstructure:"CHROMA_ADDRESS"`
	LlmAddress         string `mapstructure:"LLM_ADDRESS"`
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
	godotenv.Load("./../.env")

	config = Configuration{
		ServerAddr:         os.Getenv("SERVER_ADDR"),
		MongoConn:          os.Getenv("MONGO_CONN"),
		MongoDbName:        os.Getenv("MONGO_DB_NAME"),
		AwsS3Bucket:        os.Getenv("AWS_S3BUCKET"),
		AwsBucketRegion:    os.Getenv("AWS_BUCKET_REGION"),
		AwsBedrockRegion:   os.Getenv("AWS_BEDROCK_REGION"),
		AwsAccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AwsSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		ChromaAddress:      os.Getenv("CHROMA_ADDRESS"),
		LlmAddress:         os.Getenv("LLM_ADDRESS"),
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

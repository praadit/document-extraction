package main

import (
	"context"
	"log"
	"textract-mongo/pkg"
	"textract-mongo/pkg/controller"
	"textract-mongo/pkg/cronjob"
	"textract-mongo/pkg/integration/aws"
	"textract-mongo/pkg/integration/ollama"
	"textract-mongo/pkg/repo"
	"textract-mongo/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gin-gonic/gin"
	"github.com/jasonlvhit/gocron"
)

func main() {
	conf, _ := utils.LoadEnv()

	controller := setupController()
	server := pkg.NewServer(controller)

	job := cronjob.InitCron(controller)
	// runCron(job)

	if utils.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(server.CustomLogger())

	setupGinServer(r)
	setupRoutes(r, server)
	r.GET("/start-cron", func(ctx *gin.Context) {
		job.GetTextract()
	})

	if err := r.Run(conf.ServerAddr); err != nil {
		log.Fatal("cannot run the server:", err.Error())
	}
}

func setupController() *controller.Controller {
	// controller.InitChroma()
	customProvider := aws.AwsCredentialProvider{}

	awsConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(customProvider),
		config.WithRegion(utils.Config.AwsBucketRegion),
		func(lo *config.LoadOptions) error {
			return nil
		})
	if err != nil {
		log.Panic("failed to setup aws, err : " + err.Error())
	}

	tract := aws.InitTextract(awsConfig)
	s3 := aws.InitS3(awsConfig)
	bed := aws.InitBedrock(awsConfig, &utils.Config.AwsBedrockRegion)
	ollama := ollama.InitOllama()
	db := repo.InitDatabase(utils.Config.MongoConn, utils.Config.MongoDbName)

	return controller.NewController(db, s3, tract, bed, ollama)
}

func runCron(job *cronjob.Cron) {
	s := gocron.NewScheduler()

	s.Every(10).Minute().Do(job.GetTextract)
	s.Start()
}

func setupGinServer(r *gin.Engine) {
	r.SetTrustedProxies(nil)

	// Set up CSRF
	r.Use(gin.Recovery())
}

func setupRoutes(r *gin.Engine, server *pkg.Server) {
	r.POST("/extract", server.Controller.ExtractDocument)
	r.POST("/start-extract", server.Controller.ExtractDocumentAsync)
	r.POST("/process", server.Controller.SummarizeDocument)
	r.POST("/bedrock-process", server.Controller.BedrockSummarizeDocument)
	r.POST("/map", server.Controller.MapTable)
}

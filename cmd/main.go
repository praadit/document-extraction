package main

import (
	"log"
	"textract-mongo/pkg"
	"textract-mongo/pkg/controller"
	"textract-mongo/pkg/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	conf, _ := utils.LoadEnv()

	controller := setupController()
	server := pkg.NewServer(controller)

	if utils.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(server.CustomLogger())

	setupGinServer(r)
	setupRoutes(r, server)

	if err := r.Run(conf.ServerAddr); err != nil {
		log.Fatal("cannot run the server:", err.Error())
	}
}

func setupController() *controller.Controller {
	return controller.NewController()
}

func setupGinServer(r *gin.Engine) {
	r.SetTrustedProxies(nil)

	// Set up CSRF
	r.Use(gin.Recovery())
}

func setupRoutes(r *gin.Engine, server *pkg.Server) {
	r.POST("/extract", server.Controller.ExtractDocument)
	r.POST("/process", server.Controller.SummarizeDocument)
}

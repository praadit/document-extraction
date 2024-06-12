package pkg

import (
	"log"
	"textract-mongo/pkg/controller"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Controller *controller.Controller
}

func NewServer(controller *controller.Controller) *Server {
	return &Server{
		Controller: controller,
	}
}

func (s *Server) CustomLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now() // Start timer
		// Process request
		c.Next()

		timeStamp := time.Now() // Stop timer
		latency := timeStamp.Sub(start)
		if latency > time.Minute {
			latency = latency.Truncate(time.Second)
		}

		service := c.Request.Header.Get("x-target-service")
		log.Printf("%v | %v | %v | %v %v - source: %s", timeStamp.Format(time.RFC3339Nano), c.Writer.Status(), latency, c.Request.Method, c.Request.URL, service)
	}
}

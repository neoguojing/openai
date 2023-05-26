package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/neoguojing/log"

	"github.com/gin-gonic/gin"
	cmd "github.com/neoguojing/commander"
	"github.com/neoguojing/gormboot"
	"github.com/neoguojing/openai/server/config"
)

var (
	starter *cmd.Commander
	port    int = 8080
	logger      = log.NewLogger()
)

var Routes *gin.Engine

type Server struct {
	serv *http.Server
}

func (s *Server) Start() {
	apiKey := config.GetConfig().OpenAI.ApiKey
	Routes = GenerateGinRouter(apiKey)
	s.serv = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: Routes,
	}
	if err := s.serv.ListenAndServe(); err != nil {
		logger.Fatal(err.Error())
	}
}

func (s *Server) Stop() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.serv.Shutdown(ctx); err != nil {
		logger.Fatal(err.Error())
	}

	gormboot.Destroy()
}

func init() {
	config.GetConfig()
	starter = cmd.NewCommander()
	starter.Register(&Server{})
	// gormboot.Init()
}

func main() {
	if err := starter.Run(); err != nil {
		logger.Fatal(err.Error())
	}
}

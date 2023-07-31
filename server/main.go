package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/neoguojing/gormboot/v2"
	"github.com/neoguojing/log"

	"github.com/gin-gonic/gin"
	cmd "github.com/neoguojing/commander"
	"github.com/neoguojing/openai/config"
	"github.com/neoguojing/openai/models"
	"github.com/neoguojing/openai/role"
)

var (
	starter       *cmd.Commander
	port          int = 8080
	logger            = log.NewLogger()
	globalSession *models.SessionManager
)

var Routes *gin.Engine

type Server struct {
	serv *http.Server
}

func (s *Server) Start() {
	role.LoadRoles2DB()
	apiKey := config.GetConfig().OpenAI.ApiKey
	Routes = GenerateGinRouter(apiKey)
	globalSession = models.NewSessionManager()
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
	models.GetRecorder().Exit()
	globalSession.Close()
	chat.Destry()
	gormboot.DefaultDB.Close()

}

func init() {
	config.GetConfig()
	starter = cmd.NewCommander()
	starter.Register(&Server{})
}

func main() {
	if err := starter.Run(); err != nil {
		logger.Fatal(err.Error())
	}
}

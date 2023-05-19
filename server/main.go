package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	cmd "github.com/neoguojing/commander"
	"github.com/neoguojing/gormboot"
	"github.com/neoguojing/openai"
)

var (
	starter *cmd.Commander
	port    int = 8080
)

var Routes *gin.Engine

type Server struct {
	serv *http.Server
}

func (s *Server) Start() {
	Routes = openai.GenerateGinRouter("")
	s.serv = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: Routes,
	}
	if err := s.serv.ListenAndServe(); err != nil {
		log.Fatal(err.Error())
	}
}

func (s *Server) Stop() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.serv.Shutdown(ctx); err != nil {
		panic(err)
	}

	gormboot.Destroy()
}

func init() {

	starter = cmd.NewCommander()
	starter.Register(&Server{})
	gormboot.Init()
}

func main() {
	if err := starter.Run(); err != nil {
		log.Fatal(err.Error())
	}
}
